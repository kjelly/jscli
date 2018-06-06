# jscli
=======

Write and run simple Javascript code in common line

Introduction
------------

Write javascript code in command line, read data from stdin, and show the result you expect!
The tools is like awk. But it let you write in Javascript.

Why use jscli?

- You don't learn awk if you are good at Javascript.
- It's easy to add help function using any programming language.
- Only one executable. There are not many dependences.


Install
-------

```
$ go get github.com/kjelly/jscli
```


Example
-------

json pretty-printing

```
$ cat json.content |jscli "print(JSON.stringify(JSON.parse(stdin), null, 2))"
```

list listening address and port

```
sudo netstat -tulnp|grep LISTEN|jscli "for(i=0;i<lines.length;i+=1){print(lines[i].split(/\s+/)[3])}"
```
or

```
sudo netstat -tulnp|grep LISTEN|jscli "function showLine(i){println(i[3]);}" "matrix.map(showLine)
```
or

```
sudo netstat -tulnp|grep LISTEN|jscli "printC(matrix, 3)"
```


list listening address and port and process

```
sudo netstat -tulnp|grep LISTEN|jscli "function showLine(i){printf('%s -> %s\n', i[3], i[6]);}" "matrix.map(showLine)"
```

Javascript Builtin Function
---------------------------

Name | Explain
---- | -------
printArray | printArray
printA | printArray
printColumn | printColumn
mapIf | printColumn
mapIf | printColumn

