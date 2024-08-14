package repository

import (
	"context"
	"database/sql"
	"log"
	"project/internal/domain"
	"project/internal/repository/cache"
	"project/internal/repository/dao"
	"time"
)

/**
 * @Description
 * @Date 2024/3/1 17:49
 **/

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type CacheUsersRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewCacheUsersRepository(repo dao.UserDao, cache cache.UserCache) UserRepository {
	return &CacheUsersRepository{
		dao:   repo,
		cache: cache,
	}
}

func (repo *CacheUsersRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toDaoUser(u))

}

func (repo *CacheUsersRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}

func (repo *CacheUsersRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)
	if err == nil {
		return du, err
	}
	// err 不为nil 就要查询数据库
	//  err有两种可能
	// 1.key 不存在 说明redis正常
	// 2.访问redis 有问题 可能是网络问题，也可能是redis本身就奔溃了
	u, err := repo.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)
	// 异步写法
	// 另外回写缓存的时候忽略掉了错误，故需改善
	go func() {
		if err = repo.cache.Set(ctx, du); err != nil {
			// 网络崩了 或者redis崩了 缓存击穿
			log.Println(err)
		}

	}()

	return du, nil

}

func (repo *CacheUsersRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {

	return repo.dao.UpdateById(ctx, repo.toDaoUser(user))

}

func (repo *CacheUsersRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}
func (repo *CacheUsersRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		NickName: u.NickName,
		Birthday: time.UnixMilli(u.Birthday),
	}
}
func (repo *CacheUsersRepository) toDaoUser(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		// valid 取值为true不为空 取值为false为空
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		NickName: u.NickName,
		AboutMe:  u.AboutMe,
	}

}
