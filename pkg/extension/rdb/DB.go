package dbwrapper

import (
	"database/sql"
	"github.com/hwangtaeseung/neptune-core/pkg/common"
	"log"
	"math"
)

type HevcDB struct {
	vendor  string
	dbInfo  string
	maxConn int
	dbOff   bool
}

func NewDB(vendor string, connectString string, maxConn int) *HevcDB {
	dbOff := false
	if connectString == "" {
		dbOff = true
	}
	return &HevcDB{
		vendor:  vendor,
		dbInfo:  connectString,
		maxConn: int(math.Max(float64(maxConn), 1)),
		dbOff:   dbOff,
	}
}

func (d *HevcDB) GetDBInfo() string {
	return d.dbInfo
}

func (d *HevcDB) open() (*sql.DB, error) {
	db, err := sql.Open(d.vendor, d.dbInfo)
	if err != nil {
		log.Printf("db open error")
		return nil, err
	}
	// set max num of connection
	db.SetMaxOpenConns(d.maxConn)
	return db, nil
}

func (d *HevcDB) ExecSQLWithTx(callback func(tx *sql.Tx) (interface{}, error)) (interface{}, error) {

	return common.RetryWrapper(func() (interface{}, error) {
		if d.dbOff {
			log.Println("db off!!!")
			return []interface{}{}, nil
		}

		// open db
		db, err := d.open()
		if err != nil {
			log.Printf("db open error : %v\n", err)
			return nil, err
		}
		defer func(db *sql.DB) {
			if err := db.Close(); err != nil {
				log.Printf("db close error : %v\n", err)
			}
		}(db)

		// begin tx
		tx, err := db.Begin()
		if err != nil {
			log.Printf("transaction creation error : %+v", err)
			return nil, err
		}
		defer func(tx *sql.Tx) {
			if err := tx.Rollback(); err != nil {
				//log.Printf("rollback error! (err:%+v)", err)
			}
		}(tx)

		// execute sql callback
		if ret, err := callback(tx); err != nil {
			log.Printf("sql callback error : %+v", err)
			return nil, err
		} else {
			if err := tx.Commit(); err != nil {
				log.Printf("sql commit error : %v\n", err)
				return nil, err
			}
			// commit
			return ret, nil
		}
	}, common.DefaultMaxRetryCount)
}

type HevcSql struct {
	Sql  string
	Args []interface{}
}

func (d *HevcDB) ExecuteSQLs(sqlStatements []*HevcSql) ([]interface{}, error) {
	if results, err := d.ExecSQLWithTx(func(tx *sql.Tx) (interface{}, error) {
		var results []interface{}
		for _, sqlStatement := range sqlStatements {
			if result, err := tx.Exec(sqlStatement.Sql, sqlStatement.Args...); err != nil {
				return nil, err
			} else {
				results = append(results, result)
			}
		}
		return results, nil
	}); err != nil {
		log.Printf("exec sql error : %v\n", err)
		return nil, err
	} else {
		return results.([]interface{}), nil
	}
}
