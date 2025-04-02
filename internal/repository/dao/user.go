package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

/**
 * @Description
 * @Date 2024/3/1 18:15
 **/

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDao interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, uid int64) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{db: db}
}
func (dao *GormUserDao) Insert(ctx context.Context, u User) error {
	// 注意数据库中唯一索引冲突的问题
	now := time.Now().Unix()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突  邮箱冲突
			return ErrDuplicateEmail
		}

	}
	return err
}

func (dao *GormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err

}

func (dao *GormUserDao) FindById(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", uid).First(&u).Error
	return u, err
}

func (dao *GormUserDao) UpdateById(ctx context.Context, entity User) error {

	return dao.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).Updates(map[string]any{
		"utime":     time.Now().UnixMilli(),
		"nick_name": entity.NickName,
		"birthday":  entity.Birthday,
		"about_me":  entity.AboutMe,
	}).Error

}

func (dao *GormUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&u).Error
	return u, err

}
func (dao *GormUserDao) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id=?", openId).First(&u).Error
	return u, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// NullString  代表可以为空
	Email    sql.NullString `gorm:"unique"`
	Password string
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
	// NullString  代表可以为空
	Phone         sql.NullString `gorm:"unique"`
	NickName      string         `gorm:"type=varchar(128)"`
	Birthday      int64
	AboutMe       string         `gorm:"type=varchar(4096)"`
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
}
