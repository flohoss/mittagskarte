package mittag

func (m *Mittag) migrateModels() {
	m.orm.AutoMigrate(&Card{})
	m.orm.AutoMigrate(&Food{})
}
