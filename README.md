# myapp

Project: Covid Info. \
Tech Stack: Golang (Echo Framework) and MongoDB. \
API Doc: Swagger. \
Cloud: Heroku


Check Doc here: \
https://covid013.herokuapp.com/v1/swagger/index.html \

Set Config Var in Heroku: \
GOVERSION=go1.16 \

Set Server Conn: \
mgo "github.com/globalsign/mgo" \

server="mongodb://<username>:<password>@cluster-shard-00-00.host.mongodb.net:27017,cluster-shard-00-01.host.mongodb.net:27017,cluster-shard-00-02.host.mongodb.net:27017/<dbname>?ssl=true&authSource=admin" \

database=<db_name> \

set-up free MongoDb on Atlas or local MongoDb


