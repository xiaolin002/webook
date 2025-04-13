package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?",
			biz, id, uid, 1).
		First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, id, uid).
		First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, id).
		First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Ctime = now
	cb.Utime = now
	// 这里同样是更新和插入操作 所以要使用事务
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

// InsertLikeInfo upsert 操作在数据库层面保证了原子性，多个并发请求同时执行 upsert 操作时，数据库会正确处理冲突，避免数据重复和不一致的问题。
func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	// 由于是两个事	  所以要使用事务 都是upsert操作所以用到clause.OnConflict
	// 一是更新用户的点赞信息
	// 二是更新文章的点赞信息
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Biz:    biz,
			BizId:  id,
			Status: 1,
			Utime:  now,
			Ctime:  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id = ? AND biz=?", uid, id, biz).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			Where("biz =? AND biz_id=?", biz, id).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` - 1"),
				"utime":    now,
			}).Error
	})
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}
func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	// 插入或者更新
	/*
			避免数据重复：clause.OnConflict 会在插入记录时检查唯一索引（biz_type_id）是否冲突。
		    如果冲突，就执行更新操作，而不是插入新记录，从而避免了数据重复。
			保证计数器准确性：gorm.Expr("read_cnt + 1") 是一个 SQL 表达式，数据库会直接在服务器端执行这个操作，保证了原子性。
		   多个并发请求同时更新 read_cnt 时，每个请求都会正确地将 read_cnt 的值加 1，不会出现数据覆盖的问题。
	*/
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt` + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Cid   int64 `gorm:"index"`
	Utime int64
	Ctime int64
}

type UserLikeBiz struct {
	// 联合索引 uid-biz-bizId
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status int
	Utime  int64
	Ctime  int64
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizid, biz>
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}
