package x

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/vcraescu/go-paginator/v2"
	"github.com/vcraescu/go-paginator/v2/adapter"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type (
	RepositoryInterface interface {
		GetResult() map[string]interface{}
		CountTotal() int64
	}

	ModelInterface interface {
		GetID() string
		TableName() string
	}

	ORMRepositoryInterface interface {
		GetModel() ModelInterface
	}

	Filter struct {
		Key      interface{}
		Operator string
		Value    string
	}

	DBConfig struct {
		Server   string
		Port     uint16
		Username string
		Password string
		DBName   string
		dsn      string
	}

	Field struct {
		Key   interface{}
		Value string
	}

	Fields struct {
		Fields []*Field
		Values []interface{}
	}

	Query struct {
		Fields    Fields
		TableName string
		Surfix    string
	}

	DBHandlerInterface interface {
		Repository(repository RepositoryInterface) DBHandlerInterface
		Context(c *gin.Context) DBHandlerInterface
		Query(query Query) (map[string]interface{}, error)
		GetRepository() RepositoryInterface
	}

	ORMHandlerInterface interface {
		Repository(repository ORMRepositoryInterface) ORMHandlerInterface
		Context(c *gin.Context) ORMHandlerInterface
		Create(data ModelInterface) error
		Update(data ModelInterface) error
		Get(model ModelInterface) error
		Delete(model ModelInterface) error
		Paginate(page int, limit int, filters []Filter) ([]interface{}, error)
		GetRepository() ORMRepositoryInterface
	}

	ORMHandler struct {
		context    *gin.Context
		db         *gorm.DB
		logger     KafkaLogger
		config     DBConfig
		repository ORMRepositoryInterface
	}

	DBHandler struct {
		context    *gin.Context
		db         *sql.DB
		logger     KafkaLogger
		config     DBConfig
		repository RepositoryInterface
	}

	HTTPHandler struct {
		context *gin.Context
		logger  KafkaLogger
	}
)

func NewHTTPHandler(c *gin.Context, logger KafkaLogger) HTTPHandler {
	return HTTPHandler{
		context: c,
		logger:  logger,
	}
}

func (h HTTPHandler) Request(method string, url string, response chan<- interface{}, payload interface{}) {

}

func NewORMHandler(config DBConfig, logger KafkaLogger) ORMHandler {
	config.dsn = fmt.Sprintf("mysql://%s:%d?database=%s&charset=utf8mb4&parseTime=True&loc=Local", config.Server, config.Port, config.DBName)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password,
		config.Server, config.Port, config.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error creating connection pool: " + err.Error())
	}

	return ORMHandler{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (o *ORMHandler) Context(c *gin.Context) ORMHandlerInterface {
	o.context = c

	return o
}

func (o *ORMHandler) Repository(repository ORMRepositoryInterface) ORMHandlerInterface {
	o.repository = repository

	return o
}

func (o *ORMHandler) GetRepository() ORMRepositoryInterface {
	return o.repository
}

func (o *ORMHandler) Create(model ModelInterface) error {
	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestTrue, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "SAVE",
				}, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	var msg string
	err := o.db.Create(model).Error
	if err != nil {
		msg = err.Error()
	}

	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestFalse, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "SAVE",
				}, &msg,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return err
}

func (o *ORMHandler) Update(model ModelInterface) error {
	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestTrue, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "SAVE",
				}, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	var msg string
	err := o.db.Save(model).Error
	if err != nil {
		msg = err.Error()
	}

	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestFalse, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "SAVE",
				}, &msg,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return err
}

func (o *ORMHandler) Get(model ModelInterface) error {
	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestTrue, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "GET",
				}, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	var msg string
	err := o.db.First(model, "id = ?", model.GetID()).Error
	if err != nil {
		msg = err.Error()
	}

	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestFalse, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "GET",
				}, &msg,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return err
}

func (o *ORMHandler) Delete(model ModelInterface) error {
	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestTrue, o.config.dsn, map[string]interface{}{
					"payload":   model,
					"operation": "DELTE",
				}, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	var msg string
	err := o.db.Where("id = ?", model.GetID()).Delete(model).Error
	if err != nil {
		msg = err.Error()
	}

	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, model.TableName(),
				RequestFalse, o.config.dsn, map[string]interface{}{
					"payload":   nil,
					"operation": "DELETE",
				}, &msg,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return err
}

func (o *ORMHandler) Paginate(page int, limit int, filters []Filter) ([]interface{}, error) {
	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, o.repository.GetModel().TableName(),
				RequestTrue, o.config.dsn, map[string]interface{}{
					"payload":   filters,
					"operation": "PAGINATE",
				}, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	q := o.db.Model(o.repository.GetModel())
	for _, v := range filters {
		q.Where(fmt.Sprintf("%s %s ?", v.Key, v.Operator), v.Value)
	}

	p := paginator.New(adapter.NewGORMAdapter(q), limit)
	p.SetPage(page)

	var msg string
	var result []interface{}
	err := p.Results(&result)
	if err != nil {
		msg = err.Error()
	}

	go func() {
		err := o.logger.Send(
			o.logger.CreateMessageFromContext(
				o.context, o.config.DBName, o.repository.GetModel().TableName(),
				RequestFalse, o.config.dsn, map[string]interface{}{
					"operation": "PAGINATE",
					"payload":   result,
				}, &msg,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return result, err
}

func NewDBHandler(config DBConfig, logger KafkaLogger) DBHandler {
	config.dsn = fmt.Sprintf("sqlserver://%s:%d?database=%s&encrypt=disable", config.Server, config.Port, config.DBName)
	db, err := sql.Open("sqlserver", fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable",
		config.Username, config.Password,
		config.Server, config.Port, config.DBName,
	))
	if err != nil {
		log.Fatal("Error creating connection pool: " + err.Error())
	}

	return DBHandler{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (d *DBHandler) Repository(repository RepositoryInterface) DBHandlerInterface {
	d.repository = repository

	return d
}

func (d *DBHandler) GetRepository() RepositoryInterface {
	return d.repository
}

func (d *DBHandler) Context(c *gin.Context) DBHandlerInterface {
	d.context = c

	return d
}

func (d DBHandler) Query(query Query) (map[string]interface{}, error) {
	var fields string
	for _, v := range query.Fields.Fields {
		fields = fmt.Sprintf("%s, %s AS %s", fields, v.Key, v.Value)
	}

	sqlQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.TrimLeft(fields, ", "), query.TableName, query.Surfix)
	rows, err := d.db.Query(sqlQuery)
	if err != nil {
		go func() {
			msg := err.Error()
			err = d.logger.Send(
				d.logger.CreateMessageFromContext(
					d.context, d.config.DBName, query.TableName,
					RequestTrue, d.config.dsn, sqlQuery, &msg,
				),
			)
			if err != nil {
				fmt.Println(err.Error())
			}
		}()

		return nil, err
	}
	defer rows.Close()

	go func() {
		err := d.logger.Send(
			d.logger.CreateMessageFromContext(
				d.context, d.config.DBName, query.TableName,
				RequestTrue, d.config.dsn, sqlQuery, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	result := map[string]interface{}{}
	for rows.Next() {
		err = rows.Scan(query.Fields.Values...)
		if err != nil {
			return nil, err
		}

		for _, v := range query.Fields.Fields {
			result[v.Value] = v.Key
		}
	}

	go func() {
		err := d.logger.Send(
			d.logger.CreateMessageFromContext(
				d.context, d.config.DBName, query.TableName,
				RequestFalse, d.config.dsn, result, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return result, nil
}

func (f *Fields) Add(field Field) {
	f.Fields = append(f.Fields, &field)
	f.Values = append(f.Values, &field.Key)
}

func (f *Fields) Adds(fields ...Field) {
	wg := sync.WaitGroup{}
	wg.Add(len(fields))
	go func(me *Fields) {
		for _, f := range fields {
			me.Add(f)
			wg.Done()
		}
	}(f)

	wg.Wait()
}
