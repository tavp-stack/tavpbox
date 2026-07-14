package library

// ServiceDefinition represents a configurable service
type ServiceDefinition struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
	Image       string   `yaml:"image,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	InstallCmd  string   `yaml:"install_cmd"`
	StartCmd    string   `yaml:"start_cmd"`
	HealthCheck string   `yaml:"health_check,omitempty"`
}

// ServiceLibrary contains all built-in services
var ServiceLibrary = map[string]ServiceDefinition{
	"mariadb": {
		Name:        "mariadb",
		DisplayName: "MariaDB",
		Description: "MySQL-compatible relational database",
		Category:    "database",
		Ports:       []string{"3306:3306"},
		InstallCmd: `apt-get install -y -qq mariadb-server mariadb-client
mysql_install_db --user=root --datadir=/var/lib/mysql 2>/dev/null || true
service mariadb start 2>/dev/null || systemctl start mariadb 2>/dev/null || true
mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true`,
		StartCmd:    "service mariadb start 2>/dev/null || systemctl start mariadb 2>/dev/null || true",
		HealthCheck: "mysqladmin ping -h localhost 2>/dev/null",
	},
	"mysql": {
		Name:        "mysql",
		DisplayName: "MySQL",
		Description: "MySQL relational database",
		Category:    "database",
		Ports:       []string{"3306:3306"},
		InstallCmd: `apt-get install -y -qq mysql-server mysql-client
service mysql start 2>/dev/null || systemctl start mysql 2>/dev/null || true`,
		StartCmd:    "service mysql start 2>/dev/null || systemctl start mysql 2>/dev/null || true",
		HealthCheck: "mysqladmin ping -h localhost 2>/dev/null",
	},
	"postgres": {
		Name:        "postgres",
		DisplayName: "PostgreSQL",
		Description: "Advanced relational database",
		Category:    "database",
		Ports:       []string{"5432:5432"},
		InstallCmd: `apt-get install -y -qq postgresql postgresql-client
service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true
su - postgres -c "psql -c \"CREATE USER app WITH PASSWORD 'app' CREATEDB;\"" 2>/dev/null || true
su - postgres -c "psql -c \"CREATE DATABASE app OWNER app;\"" 2>/dev/null || true`,
		StartCmd:    "service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true",
		HealthCheck: "pg_isready -h localhost 2>/dev/null",
	},
	"mongodb": {
		Name:        "mongodb",
		DisplayName: "MongoDB",
		Description: "NoSQL document database",
		Category:    "database",
		Ports:       []string{"27017:27017"},
		InstallCmd: `curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg 2>/dev/null
echo "deb [ signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] http://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" > /etc/apt/sources.list.d/mongodb-org-7.0.list
apt-get update -qq && apt-get install -y -qq mongodb-org
systemctl start mongod 2>/dev/null || service mongod start 2>/dev/null || true`,
		StartCmd:    "systemctl start mongod 2>/dev/null || service mongod start 2>/dev/null || true",
		HealthCheck: "mongosh --eval 'db.runCommand({ping:1})' 2>/dev/null",
	},
	"redis": {
		Name:        "redis",
		DisplayName: "Redis",
		Description: "In-memory cache & queue",
		Category:    "cache",
		Ports:       []string{"6379:6379"},
		InstallCmd: `apt-get install -y -qq redis-server
systemctl start redis-server 2>/dev/null || service redis-server start 2>/dev/null || true`,
		StartCmd:    "systemctl start redis-server 2>/dev/null || service redis-server start 2>/dev/null || true",
		HealthCheck: "redis-cli ping 2>/dev/null",
	},
	"memcached": {
		Name:        "memcached",
		DisplayName: "Memcached",
		Description: "Distributed memory cache",
		Category:    "cache",
		Ports:       []string{"11211:11211"},
		InstallCmd: `apt-get install -y -qq memcached
systemctl start memcached 2>/dev/null || service memcached start 2>/dev/null || true`,
		StartCmd:    "systemctl start memcached 2>/dev/null || service memcached start 2>/dev/null || true",
		HealthCheck: "echo 'stats' | nc -w 1 localhost 11211 2>/dev/null | head -1",
	},
	"mailpit": {
		Name:        "mailpit",
		DisplayName: "Mailpit",
		Description: "Email testing SMTP server",
		Category:    "mail",
		Ports:       []string{"8025:8025", "1025:1025"},
		InstallCmd: `curl -sL https://github.com/axllent/mailpit/releases/latest/download/mailpit_linux_amd64.tar.gz | tar xz -C /usr/local/bin/
cat > /etc/systemd/system/mailpit.service <<'EOF'
[Unit]
Description=Mailpit
After=network.target
[Service]
ExecStart=/usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025
Restart=always
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload && systemctl enable mailpit && systemctl start mailpit 2>/dev/null || true`,
		StartCmd:    "systemctl start mailpit 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost:8025/api/v1/info 2>/dev/null | head -1",
	},
	"mailhog": {
		Name:        "mailhog",
		DisplayName: "MailHog",
		Description: "Email testing tool",
		Category:    "mail",
		Ports:       []string{"8025:8025", "1025:1025"},
		InstallCmd: `curl -sL https://github.com/mailhog/MailHog/releases/latest/download/MailHog_linux_amd64 -o /usr/local/bin/mailhog
chmod +x /usr/local/bin/mailhog
cat > /etc/systemd/system/mailhog.service <<'EOF'
[Unit]
Description=MailHog
After=network.target
[Service]
ExecStart=/usr/local/bin/mailhog
Restart=always
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload && systemctl enable mailhog && systemctl start mailhog 2>/dev/null || true`,
		StartCmd:    "systemctl start mailhog 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost:8025/api/v2/messages 2>/dev/null | head -1",
	},
	"phpmyadmin": {
		Name:        "phpmyadmin",
		DisplayName: "phpMyAdmin",
		Description: "Database admin UI",
		Category:    "admin",
		Ports:       []string{"8080:80"},
		InstallCmd: `apt-get install -y -qq phpmyadmin
ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true`,
		StartCmd:    "service apache2 start 2>/dev/null || systemctl start apache2 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost/pma 2>/dev/null | head -1",
	},
	"adminer": {
		Name:        "adminer",
		DisplayName: "Adminer",
		Description: "Lightweight database manager",
		Category:    "admin",
		Ports:       []string{"8080:80"},
		InstallCmd: `mkdir -p /var/www/html/adminer
curl -sL https://www.adminer.org/latest.php -o /var/www/html/adminer/index.php
curl -sL https://www.adminer.org/download/v5.4.4/designs/haeckel/adminer.css -o /var/www/html/adminer/adminer.css
chmod 644 /var/www/html/adminer/index.php /var/www/html/adminer/adminer.css`,
		StartCmd:    "service nginx start 2>/dev/null || systemctl start nginx 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost/adminer/ 2>/dev/null | head -1",
	},
	"elasticsearch": {
		Name:        "elasticsearch",
		DisplayName: "Elasticsearch",
		Description: "Search & analytics engine",
		Category:    "search",
		Ports:       []string{"9200:9200"},
		InstallCmd: `curl -fsSL https://artifacts.elastic.co/GPG-KEY-elasticsearch | gpg --dearmor -o /usr/share/keyrings/elasticsearch-keyring.gpg 2>/dev/null
echo "deb [signed-by=/usr/share/keyrings/elasticsearch-keyring.gpg] https://artifacts.elastic.co/packages/8.x/apt stable main" > /etc/apt/sources.list.d/elastic-8.x.list
apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq elasticsearch
systemctl start elasticsearch 2>/dev/null || service elasticsearch start 2>/dev/null || true`,
		StartCmd:    "systemctl start elasticsearch 2>/dev/null || service elasticsearch start 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost:9200 2>/dev/null | head -1",
	},
	"rabbitmq": {
		Name:        "rabbitmq",
		DisplayName: "RabbitMQ",
		Description: "Message broker",
		Category:    "queue",
		Ports:       []string{"5672:5672", "15672:15672"},
		InstallCmd: `apt-get install -y -qq rabbitmq-server
systemctl start rabbitmq-server 2>/dev/null || service rabbitmq-server start 2>/dev/null || true
rabbitmq-plugins enable rabbitmq_management 2>/dev/null || true`,
		StartCmd:    "systemctl start rabbitmq-server 2>/dev/null || service rabbitmq-server start 2>/dev/null || true",
		HealthCheck: "rabbitmqctl status 2>/dev/null | head -1",
	},
	"beanstalkd": {
		Name:        "beanstalkd",
		DisplayName: "Beanstalkd",
		Description: "Simple work queue",
		Category:    "queue",
		Ports:       []string{"11300:11300"},
		InstallCmd: `apt-get install -y -qq beanstalkd
systemctl start beanstalkd 2>/dev/null || service beanstalkd start 2>/dev/null || true`,
		StartCmd:    "systemctl start beanstalkd 2>/dev/null || service beanstalkd start 2>/dev/null || true",
		HealthCheck: "echo 'stats' | nc -w 1 localhost 11300 2>/dev/null | head -1",
	},
	"apache": {
		Name:        "apache",
		DisplayName: "Apache",
		Description: "Apache HTTP web server",
		Category:    "webserver",
		Ports:       []string{"80:80"},
		InstallCmd: `apt-get install -y -qq apache2
systemctl start apache2 2>/dev/null || service apache2 start 2>/dev/null || true`,
		StartCmd:    "systemctl start apache2 2>/dev/null || service apache2 start 2>/dev/null || true",
		HealthCheck: "curl -s http://localhost 2>/dev/null | head -1",
	},
	"varnish": {
		Name:        "varnish",
		DisplayName: "Varnish",
		Description: "HTTP cache / reverse proxy",
		Category:    "cache",
		Ports:       []string{"80:80"},
		InstallCmd: `apt-get install -y -qq varnish
systemctl start varnish 2>/dev/null || service varnish start 2>/dev/null || true`,
		StartCmd:    "systemctl start varnish 2>/dev/null || service varnish start 2>/dev/null || true",
		HealthCheck: "varnishadm status 2>/dev/null",
	},
}

// GetService returns a service definition by name
func GetService(name string) (ServiceDefinition, bool) {
	svc, ok := ServiceLibrary[name]
	return svc, ok
}

// ListServices returns all available services
func ListServices() []string {
	var names []string
	for name := range ServiceLibrary {
		names = append(names, name)
	}
	return names
}
