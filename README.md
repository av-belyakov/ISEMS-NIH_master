Application ISEMS-NIH master, v1.7.6
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
