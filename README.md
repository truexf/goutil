# goutil
golang util
常用函数和工具类

## jsonexp
一个json表达式组件，赋予json的语义化的能力

## linked_list
双向链表， 支持线程安全和非安全，最大化满足性能要求

## queue
分段式slice，在内存分配和性能之间取得平衡，类似c++ stl的dequeue

## ring_queue
环形队列

## safe_rand
线程安全的随机数生成器

## token_bucket
令牌桶，典型的使用场景是用于流量控制和调节

## mysqlutils
mysql的工具类函数，最主要的是实现对sql语句的参数化处理，避免sql注入的同时方便sql的日志化(将参数实例代入sql，产生明确的sql)

## goutils.go
常用的通用化函数和类
