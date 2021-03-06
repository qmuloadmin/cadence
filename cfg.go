package cadence

/*
cfg library manages the reading and writing of the configuration file
Defined herein are the types for containing service configuration, namely
Zone and Conf, which contains Zones. Also defined here is the function to
return a new Conf object from the configuration file, and the bound Conf method
to save the current Conf to file.
*/

func load() Conf {
	local_config := Conf{}
	local_config.new()
	return local_config
}

func (c *Conf) save() {

}

func (c *Conf) new() {
	c.self = "host.example.com"
	c.zones = []Zone{Zone{}}
	c.clientPort = 6442
	c.hostPort = 3002
	c.autonomous = false
	c.my_zone = &c.zones[0]
	c.log_file = "cadence.log"
	c.environment = make(map[string]string)
}

func (c *Conf) is_zero() bool {
	if c.clientPort == 0 && c.self == "" && c.hostPort == 0 {
		return true
	}
	return false
}
