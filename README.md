# orm

## 起因
*    每次都存取map太麻烦了...

### 怎么用？

#### 初始化
    var client orm.Client //先声明一个client 嗯，可以是全局的说

    func init(){
    	c, err := mysql.NewDefault("链接字符串")
    	//初始化这个client,先引入github.com/go-sql-driver/mysql驱动
	    if err != nil {

		    c.Close()
    
		    return
	    }
	    client = c 
    }

#### 增删改

*    用的都是Execute方法，返回sql.Result,其中包含2个方法,LastInsertId可以获得插入后最新行的id,RowsAffected可以获得此次操作影响的行数

        func Insert(){
            sql := "INSERT INTO test(t1,t1) VAULES(?,?)"
        
	        result, err := client.Execute(sql, 1, 2)
	        if err != nil {
	        	t.Error(err)
	        }
	        fmt.Println(result.LastInsertId())
        }

        func Update(){
            sql := "UPDATE test SET t1=? WHERE t2=?"
        
	        result, err := client.Execute(sql, 1, 2)
	        if err != nil {
	        	t.Error(err)
	        }
	        fmt.Println(result.RowsAffected())
        }

        func Delete(){
            sql := "DELETE FROM test"
        
	        result, err := client.Execute(sql, 1, 2)
	        if err != nil {
	        	t.Error(err)
	        }
	        fmt.Println(result.RowsAffected())
        }

#### 查

*    使用Query方法,只返回了Error,查询结果会拼装到你传入的变量里。
     参数v支持slice, struct,map[string]interface{},基础类型。


        func SelectOne(){
            sql := "SELECT 10 FROM test"
           
            var v int64 = -10
           
            err := client.Query(&v, sql)
            if err != nil {
            	t.Error(err)
            }
            fmt.Println(v)
        }

        type T struct{
             Id string
             Name string
        }

        func SelectStruct(){
            //列名必须和T中的属性名一致
            sql := "SELECT id AS Id,name AS Name FROM test" 
           
            v:=T{}
           
            err := client.Query(&v, sql)
            if err != nil {
            	t.Error(err)
            }
            fmt.Println(v)
        }

        func SelectMap(){
            //Id 和 Name 作为map的key值
            sql := "SELECT id AS Id,name AS Name FROM test"
           
            v:=make(map[string]interface{})
           
            err := client.Query(&v, sql)
            if err != nil {
            	t.Error(err)
            }
            fmt.Println(v)
        } 

        func SelectSliceMap(){
            //Id 和 Name 作为map的key值
            sql := "SELECT id AS Id,name AS Name FROM test"
           
            var v []map[string]interface{}
           
            err := client.Query(&v, sql)
            if err != nil {
            	t.Error(err)
            }
            fmt.Println(v)
        }  

        func SelectSliceStruct(){
            //Id 和 Name 作为map的key值
            sql := "SELECT id AS Id,name AS Name FROM test"
           
            var v []T
           
            err := client.Query(&v, sql)
            if err != nil {
            	t.Error(err)
            }
            fmt.Println(v)
        }        

