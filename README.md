Application ISEMS-NIH master, v1.6.5
Information Security Event Management System Network Interaction Handler (ISEMS-NIH)

Сервер сетевого взаимодействия с территориально удаленными источниками ISEMS-NIH slave.
Применяется для транслирования команд и скачивания файлов.

Настройка СУБД MongoDB
use <имя_БД>

db.createUser({
	user:"", 
	pwd:"", 
	roles:[{role:"readWrite", db:"isems-nih"}], 
	authenticationRestrictions: [{
	    clientSource:[""], 
	    serverAddress:[""]
	}]
})

    "connectionDB": {
        "socket": false,
        "host": "127.0.0.1",
        "port": 37017,
        "user": "module-isems-nih",
        "password": "tkovomfh&ff93",
        "nameDB": "isems-nih",
        "unixSocketPath": "../../../../tmp/mongodb.sock"
    },