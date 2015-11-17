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