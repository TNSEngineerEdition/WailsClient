package consts

// passengers
const (
	// TODO: 15 is only for a presentation purposes, change to 30 (or whatever value, like 45) later
	MAX_WAITING_TIME = 15 * 60 // 15 min
	// TODO: 1 is only for a presentation purposes, change to 6 (or whatever value) later
	MAX_PASSENGERS_CHANGE_RATE = 1      // max number of passengers per second
	TRAM_CHANGE_TIME           = 2 * 60 // 2 min
	TRAM_CHANGE_PROBABILITY    = 0.5
	DESPAWN_TIME_OFFSET        = 5 * 60 // 5 min
)
