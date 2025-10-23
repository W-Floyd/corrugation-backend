package backend

import "log"

func init() {
	var err error

	log.Println("init backend")
	err = ConnectDB("./db.sqlite")
	if err != nil {
		log.Fatalln(err)
	}
	err = InitAndMigrateDB()
	if err != nil {
		log.Fatalln(err)
	}
}
