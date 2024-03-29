## JSON表达式 
JSON表达式一个简单优雅而又功能强大的“语义化”的配置语言，可用于各种系统的配置中，让原本静态的配置参数活起来。给你带来全新的思维方式。 
JSON表达式组是一个由 n个“JSON表达式”组成的json数组，如下: 
```
{
	"my_json_exp_group": [
		//JSON表达式1
		[

		],
		//JSON表达式2
		[

		]
		//...
	]
}
```
一个“JSON表达式”的格式如下： 
它（“JSON表达式”）是一个json数组，数组的最后一个成员是赋值表达式，其他成员是条件表达式，当所有的条件表达式都成立(返回 真（true）)的时 候， 赋值表达式被执行。如果赋值表达式本身是一个数组，表示多重赋值。 
条件表达式和赋值表达式，是一个包含三个成员的形式，第一个成员为系统变量，第二个成员为比较运算符或赋值运算符，第三个成员是运算对象，比如：
```
[
    ["$hour","=","10"],
    ["$resp.source_icon","=",""],
    ["$resp.source_icon","=","http://xxxxx.jpg"]
]
```
该"JSON表达式"表达的意思是，如果现在时钟在上午十点区间内，且$resp.source_icon的值为空白，则对$resp.source_icon赋值为http://xxxx.jpg。 

若整个"JSON表达式"组中包含不止一个"JSON表达式"，则是按照编写的顺序，从上到下依次执行各个"JSON表达式"的。遇到$break=1,则不再往下执行。 

如果"JSON表达式“只有一个成员，这个成员就是赋值表达式。也就是没有条件表达式时，表示无条件执行该赋值表达式。
```
[
    //由于没有条件表达式，以下赋值表达式无条件执行
    ["$resp.source_icon","=","http://xxxxx.jpg"]
]
```

如果赋值表达式本身是一个数组，表示多重赋值。 
以下举一个多重赋值的例子：
```
[
    ["$resp.hour","=","10"],
    ["$resp.source_icon","=",""],
    //以下赋值表达式是一个多重赋值表达式，对两个系统变量进行了赋值。
    [ 
        ["$resp.source_icon","=","http://xxxxx.jpg"],
        ["$resp.source","=","ad"]
    ]
]
```
### 管道
管道支持对变量进行管道化处理  
格式： $varName[|pipeLineFunction1[|pipeLineFunction2[|...]]]  
举例：   
```
[
    ["$my_var|md5|fnv32",">", 100000],
    ["$my_var","=", 100000]
]
```
表示“当变量$my_var的md5哈希的fnv32哈希值>100000时，将$my_var的值设置为100000”

### 宏
表达式的右值支持宏替换   
宏格式：{{$variant}}  
举例：  
```
[
    ["$my_var","=", "now: {{$date}}"]
]
```
最终$my_var将被赋值为 now:2021-11-23


### 系统变量
表达式中的变量命名必须以$开头，且必须通过Dictionary.RegisterVar进行注册后才可以使用。预定义变量如下： 
* $datetime	string	yyyy-mm-dd hh:nn:ss 
* $date	string	yyyy-mm-dd 
* $time	string	hh:nn:ss 
* $stime	string	hh:nn 
* $year	string	yyyy 
* $month	string	mm 
* $day	string	dd 
* $hour	string	hh 
* $minute	string	nn 
* $second	string	ss 
* $iyear	int	当前4位年份 
* $imonth	int	当前月份整数 
* $iday	int	当前天1~30(,28,29,31)整数 
* $ihour	int	当前小时0~23 
* $iminute	int	当前分钟0~59 
* $isecond	int	当前秒0~59 
* $rand	int	1-100的随机数 
* $break int 当值为1时，终止当前条件表达式组的执行

### 条件(比较)运算符 
* \>   大于 
* \>=  大于等于 
* <   小于 
* <=  小于等于 
* =   等于 
* <>  不等于 
* !=  不等于 
* between	在区间，例如： 
[“$ihour”,”between”,”5,10”]表示如果当前时间在5到10点之间
* ^between	between的反义词 
* in	在列表中例如： 
* [“$hour”,”in”,”05,06,07,10”]
* not in	In的反义词，不在列表中 
* has	集合操作符，左值和右值都是以逗号分隔开的集合。 包含，例如 
[“$req.mimes”,”has”,”jpg,png”] 
* any	集合操作符，左值和右值都是以逗号分隔开的集合。包含逗号分隔开的一组字符串中的任意一个 
* none	集合操作符，左值和右值都是以逗号分隔开的集合。any的反义词，一个都不包含 
* ~	包含部分字符串 
* ^~	~的反义词，不包部分字符串 
* ~*	头部匹配，比如： 
* abcdefg ~* abc 返回真 
* abcdefg ~* cde 返回假 
* *~	尾部匹配，比如： 
abcdefg ~* abc 返回假 
abcdefg ~* efg 返回真 
* ^~*	~*的反义词 
* ^*~	*~的反义词 
* cv	包含逗号分隔开的部分字符串中的一个或多个，例如： 
[“$resp.title”,”cv”,”整形,医疗,美容,减肥”] 
* ^cv	cv的反义词 

### 系统预定义管道函数
* len	int	返回入参的字符个数
* upper	string	返回入参的英文大写
* lower	string	返回入参的英文小写
* md5 string 返回入参的md5哈希hex值，小写
* MD5 string 返回入参的md5哈希hex值，大写
* fnv32 uint32 返回入参的fnv32哈希值
* fnv64 uint64 返回入参的fnv64哈希值

### 赋值操作符
* =	赋值
* +=	给自己加上某个值
* -=	给自己减去某个值
* *=	给自己乘以某个值，如果是字符串，则对字符串进行复制n遍添加到末尾
* /=	给自己除以某个值
* %=	给自己除模

