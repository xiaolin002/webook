package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"time"
)

/**
 * @Description
 * @Date 2024/3/3 18:55
 **/

func InitTables(db *gorm.DB) error {

	return db.AutoMigrate(&User{})

}

// mongodb 初始化
func InitCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	col := mdb.Collection("articles")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	return err
}
