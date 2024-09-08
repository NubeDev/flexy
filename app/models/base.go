package model

import (
	"database/sql/driver"
	"fmt"
	"github.com/NubeDev/flexy/utils/setting"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

var (
	db          *gorm.DB
	TablePrefix string
)

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

type PaginateStruct struct {
	PageNum  int
	PageSize int
}

type BaseModel struct {
	ID uint `gorm:"primary_key" json:"id"`
	//UUID string `gorm:"primary_key" json:"uuid"`
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"public" json:"deleted_at"`
}

type BaseModelNoId struct {
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"public" json:"deleted_at"`
}

func Setup() {
	var err error
	var dsn string

	var newLogger logger.Interface

	// Enable SQL logs if configured
	if setting.DatabaseSetting.EchoSql {
		newLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      false,       // True - Don't include params in the SQL log
				Colorful:                  true,
			},
		)
	}

	if setting.DatabaseSetting.Type == "" {
		setting.DatabaseSetting.Type = "sqlite"
	}

	// Switch between SQLite and MySQL based on the database type in the configuration
	if setting.DatabaseSetting.Type == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=20s",
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.Name,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	} else if setting.DatabaseSetting.Type == "sqlite" {
		dsn = fmt.Sprintf("%s.db", "database") // SQLite DSN
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: newLogger})
	}

	if db == nil {
		log.Fatalf("Database setup failed: %s type: %s", "failed to initialise the database", setting.DatabaseSetting.Type)
	}

	// Check for any errors in opening the connection
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}

	// Setup connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Optional: Set table prefix if needed
	TablePrefix = setting.DatabaseSetting.TablePrefix

	// Auto-migrate your models
	db.AutoMigrate(
		&Report{},
		&Auth{},
		&JwtBlacklist{},
		&Role{},
		&Menu{},
		&Host{},
	)
}

func DBClose() {
	sql, _ := db.DB()
	sql.Close()
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t JSONTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte(`null`), nil
	}
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

/*func (v BaseModel) BeforeCreate(scope *gorm.Scope) error {
   scope.SetColumn("created_at", time.Now())
   scope.SetColumn("updated_at", time.Now())
   return nil
}

func (v BaseModel) BeforeUpdate(scope *gorm.Scope) error {
   scope.SetColumn("updated_at", time.Now())
   return nil
}*/

func SoftDelete(tableStruct interface{}) (error, int64) {
	log.Println(tableStruct)
	res := db.Model(tableStruct).Update("deleted_at", time.Now())
	if err := res.Error; err != nil {
		return err, 0
	}
	return nil, res.RowsAffected
}

func Update(tableStruct interface{}, wheres map[string]interface{}, updates map[string]interface{}) (error, int64) {
	res := db.Model(tableStruct).Where(wheres).Updates(updates)
	if err := res.Error; err != nil {
		return err, 0
	}
	return nil, res.RowsAffected
}

func GetTotal(tableStruct interface{}, where map[string]interface{}) (int64, error) {
	var count int64
	var dbData, err = BuildCondition(db, where)

	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	if err := dbData.Model(tableStruct).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func BuildCondition(d *gorm.DB, where map[string]interface{}) (*gorm.DB, error) {
	for key, value := range where {
		conditionKey := strings.Split(key, " ")
		if len(conditionKey) != 2 {
			return nil, fmt.Errorf("the condition format for building the map is incorrect")
		}

		field := conditionKey[0]
		operator := conditionKey[1]

		switch operator {
		case "=":
			d = d.Where(fmt.Sprintf("%s = ?", field), value)
		case ">":
			d = d.Where(fmt.Sprintf("%s > ?", field), value)
		case ">=":
			d = d.Where(fmt.Sprintf("%s >= ?", field), value)
		case "<":
			d = d.Where(fmt.Sprintf("%s < ?", field), value)
		case "<=":
			d = d.Where(fmt.Sprintf("%s <= ?", field), value)
		case "in":
			d = d.Where(fmt.Sprintf("%s IN (?)", field), value)
		case "like":
			d = d.Where(fmt.Sprintf("%s LIKE ?", field), value)
		case "<>", "!=":
			d = d.Where(fmt.Sprintf("%s != ?", field), value)
		case "is":
			if value == nil {
				d = d.Where(fmt.Sprintf("%s IS NULL", field))
			} else {
				d = d.Where(fmt.Sprintf("%s = ?", field), value)
			}
		default:
			return nil, fmt.Errorf("unsupported operator：%s", operator)
		}
	}

	return d, nil
}
