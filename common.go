package orm

type sqlParameter map[string]interface{}

func (this sqlParameter) Set(key string, value interface{}) SqlParameter {
	this[key] = value
	return this
}

func NewSqlParameter() SqlParameter {
	return sqlParameter{}
}
