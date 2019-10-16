Application ISEMS-NIH master, v0.32
Information Security Event Management System Network Interaction Handler (ISEMS-NIH)

Сервер сетевого взаимодействия с территориально удаленными источниками ISEMS-NIH slave.
Применяется для транслирования команд и скачивания файлов.

Настройка СУБД MongoDB
use <имя_БД>

db.createUser({
	user:"module-isems-nih", 
	pwd:"tkovomfh&ff93", 
	roles:[{role:"readWrite", db:"isems-nih"}], 
	authenticationRestrictions: [{
	    clientSource:["127.0.0.1","195.161.164.38","81.177.34.205"], 
	    serverAddress:["127.0.0.1","195.161.164.38","81.177.34.205"]
	}]
})