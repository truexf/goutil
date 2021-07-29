## JSON表达式
广告主/广告位/发布商/全局配置中都有一个json底层元素，其key可能是filter,ad-filter,cond，cond0,cond1...等等,其值是一个JSON表达式组。JSON表达式组是一个由 n个“JSON表达式”组成的json数组，如下: 
```
{
	"ad - filter": [
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
    ["$resp.hour","=","10"],
    ["$resp.source_icon","=",""],
    ["$resp.source_icon","=","http://xxxxx.jpg"]
]
```
该"JSON表达式"表达的意思是，如果现在时钟在上午十点区间内，且$resp.source_icon的值为空白，则对$resp.source_icon赋值为http://xxxx.jpg。 

若整个"JSON表达式"组中包含不止一个"JSON表达式"，则是按照编写的顺序，从上到下依次执行各个"JSON表达式"的。遇到$give_up=1,则不再往下执行。 

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

### 系统变量
名称	数据类型	说明 
$datetime	string	yyyy-mm-dd hh:nn:ss 
$date	string	yyyy-mm-dd 
$time	string	hh:nn:ss 
$stime	string	hh:nn 
$year	string	yyyy 
$month	string	mm 
$day	string	dd 
$hour	string	hh 
$minute	string	nn 
$second	string	ss 
$iyear	int	当前4位年份 
$imonth	int	当前月份整数 
$iday	int	当前天1~30(,28,29,31)整数 
$ihour	int	当前小时0~23 
$iminute	int	当前分钟0~59 
$isecond	int	当前秒0~59 
$rand	int	1-100的随机数 

### 条件(比较)运算符 
* >   大于 
* >=  大于等于 
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

### 变量后缀
变量后缀是在变量后面以英文圆点追加的特定意义的后缀，可以对变量执行对应函数，其结果为函数执行结果，比如： 
$resp.landing.len
其中.len是$resp.landing的后缀，其功能是返回该变量的字符个数 
* .len	int	返回变量值的字符个数
* .upper	string	返回变量值的英文大写
* .lower	string	返回变量值的英文小写

### 赋值操作符
* =	赋值
* +=	给自己加上某个值
* -=	给自己减去某个值
* *=	给自己乘以某个值，如果是字符串，则对字符串进行复制n遍添加到末尾
* /=	给自己除以某个值
* %=	给自己除模

