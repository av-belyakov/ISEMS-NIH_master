package configure

//ChannelCollectionCoreApp коллекция каналов для coreAppRoute
type ChannelCollectionCoreApp struct {
	OutCoreChanDB, InCoreChanDB   chan *MsgBetweenCoreAndDB
	OutCoreChanAPI, InCoreChanAPI chan *MsgBetweenCoreAndAPI
	OutCoreChanNI, InCoreChanNI   chan *MsgBetweenCoreAndNI
	ChanDropNI                    chan string
}

//MsgWsTransmission содержит информацию для передачи подключенному источнику
type MsgWsTransmission struct {
	DestinationHost string
	Data            *[]byte
}

// MsgBetweenCoreAndNI используется для взаимодействия между ядром приложения и модулем сет. взаимодействия
// TaskID - ID задачи
// ClientName - имя клиента, используется для управления источниками
// Section:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'error notification'
//  - 'message notification'
// Command:
//  - для source_control:
//		* 'load list'
// 		* 'update list' (для каждого источника свое возможно действие например,
//			один источник нужно добавить 'add', другой удалить 'del', а третий
// 			перезапустить 'reconnect')
// 		* 'keep list sources in database' (сохранить список источников в БД)
// 		* 'send list sources to client api'
// 		* 'change connection status source'
// 		* 'telemetry'
//  - для filtration_control:
//		* 'start'
//		* 'stop'
//		* ''
//  - для download_control:
//		* 'start'
//		* 'stop'
//  - 'error notification'
//      * 'send client API'
//      * 'no send client API'
//  - 'message notification'
//      * 'send client API'
//      * 'no send client API'
type MsgBetweenCoreAndNI struct {
	TaskID          string
	ClientName      string
	Section         string
	Command         string
	SourceID        int
	AdvancedOptions interface{}
}

//MsgBetweenCoreAndAPI используется для взаимодействия между ядром приложения и модулем API приложения
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - API module/Core module
// MsgType:
//  - information
//  - command
// MsgSection:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'information search control'
//  - 'source telemetry info'
//  - 'error notification'
//  - 'message notification'
// IDClientAPI - уникальный идентификатор клиента API
type MsgBetweenCoreAndAPI struct {
	MsgGenerator string
	MsgRecipient string
	IDClientAPI  string
	ClientName   string
	ClientIP     string
	MsgJSON      interface{}
}

//MsgBetweenCoreAndDB используется для взаимодействия между ядром и модулем взаимодействия с БД
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - DB module/Core module (источник сообщения)
// MsgRecipient - API module/NI module/DB module/Core module (получатель сообщения)
// MsgSection:
//  - 'source control'
//  - 'source telemetry'
//  - 'filtration'
//  - 'download'
//  - 'information search results'
//  - 'error notification'
//  - 'message notification'
//  - 'information search'
// Insturction:
//  - insert
//  - find
//  - find_all
//  - update
//  - delete
// IDClientAPI - уникальный идентификатор клиента API
// TaskID - уникальный идентификатор задачи присвоенный ядром приложения
type MsgBetweenCoreAndDB struct {
	MsgGenerator    string
	MsgRecipient    string
	MsgSection      string
	Instruction     string
	IDClientAPI     string
	TaskID          string
	TaskIDClientAPI string
	AdvancedOptions interface{}
}

// --- AdvancedOptions взависимости от типов сообщений ---

//ErrorNotification содержит информацию об ошибке
// SourceReport - DB module/NI module/API module
// ErrorBody - тело ошибки (stack trace)
// HumanDescriptionError - сообщение для пользователя
// Sources - срез ID источников связанных с данным сообщением
type ErrorNotification struct {
	SourceReport, HumanDescriptionError string
	ErrorBody                           error
	Sources                             []int
}

//MessageNotification содержит информационное сообщение о выполненном действии
// SourceReport - DB module/NI module/API module
// Section - раздел к которому относится действие
// TypeActionPerformed - тип выполненного действия
// CriticalityMessage - критичность сообщения ('info', 'success', 'warning', 'danger')
// HumanDescriptionError - сообщение для пользователя
// Sources - срез ID источников связанных с данным сообщением
type MessageNotification struct {
	SourceReport                 string
	Section                      string
	TypeActionPerformed          string
	CriticalityMessage           string
	HumanDescriptionNotification string
	Sources                      []int
}
