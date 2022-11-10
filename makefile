target:
	for number in 1 2 3; do \
		sqlite3 -init db_migrations/$$number* gym_routine.db .quit ; \
	done	