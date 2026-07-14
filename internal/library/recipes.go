package library

// RecipeDefinition represents a project recipe
type RecipeDefinition struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Description string   `yaml:"description"`
	Image       string   `yaml:"image"`
	Webroot     string   `yaml:"webroot"`
	Services    []string `yaml:"default_services"`
	InstallCmd  string   `yaml:"install_cmd"`
	Env         map[string]string `yaml:"env,omitempty"`
}

// RecipeLibrary contains all built-in recipes
var RecipeLibrary = map[string]RecipeDefinition{
	"tavp": {
		Name:        "tavp",
		DisplayName: "TAVP Stack",
		Description: "Tailwind + Alpine.js + Volt + Phalcon (PHP 8.3)",
		Image:       "docker.io/library/ubuntu:24.04",
		Webroot:     "public",
		Services:    []string{"mariadb", "redis", "mailpit"},
		InstallCmd: `apt-get update -qq && apt-get install -y -qq \
  nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml \
  php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql \
  php8.3-pgsql php8.3-redis php8.3-sqlite3 php8.3-gd \
  php-pear php8.3-dev gcc make composer curl wget git unzip

# Install Phalcon
pecl channel-update pecl.php.net 2>/dev/null
pecl install phalcon 2>/dev/null
echo "extension=phalcon.so" > /etc/php/8.3/mods-available/phalcon.ini
phpenmod phalcon 2>/dev/null || true

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq nodejs

# Configure nginx
cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html/public;
    index index.php index.html;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX

# Start services
systemctl start php8.3-fpm nginx 2>/dev/null || service php8.3-fpm start && service nginx start`,
		Env: map[string]string{
			"APP_ENV": "local",
		},
	},
	"laravel": {
		Name:        "laravel",
		DisplayName: "Laravel",
		Description: "Laravel PHP framework",
		Image:       "docker.io/library/ubuntu:24.04",
		Webroot:     "public",
		Services:    []string{"mariadb", "redis", "mailpit"},
		InstallCmd: `apt-get update -qq && apt-get install -y -qq \
  nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml \
  php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql \
  php8.3-pgsql php8.3-redis php8.3-sqlite3 php8.3-gd \
  php-pear php8.3-dev gcc make composer curl wget git unzip

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq nodejs

# Configure nginx
cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html/public;
    index index.php;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX

# Start services
systemctl start php8.3-fpm nginx 2>/dev/null || service php8.3-fpm start && service nginx start`,
		Env: map[string]string{
			"APP_ENV":  "local",
			"APP_URL":  "http://myapp.tavp.my.id",
			"DB_HOST":  "localhost",
			"DB_PORT":  "3306",
			"DB_DATABASE": "app",
			"DB_USERNAME": "app",
			"DB_PASSWORD": "app",
		},
	},
	"php": {
		Name:        "php",
		DisplayName: "PHP",
		Description: "Generic PHP application",
		Image:       "docker.io/library/ubuntu:24.04",
		Webroot:     "public",
		Services:    []string{"mariadb", "redis"},
		InstallCmd: `apt-get update -qq && apt-get install -y -qq \
  nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml \
  php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql \
  php8.3-pgsql php8.3-redis php8.3-sqlite3 php8.3-gd \
  composer curl wget git unzip

# Configure nginx
cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html;
    index index.php index.html;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX

# Start services
systemctl start php8.3-fpm nginx 2>/dev/null || service php8.3-fpm start && service nginx start`,
	},
	"node": {
		Name:        "node",
		DisplayName: "Node.js",
		Description: "Node.js application",
		Image:       "docker.io/library/node:20-alpine",
		Webroot:     ".",
		Services:    []string{"redis"},
		InstallCmd: `apk add --no-cache nginx
npm install -g yarn pnpm

# Configure nginx
cat > /etc/nginx/http.d/default.conf <<'NGINX'
server {
    listen 80;
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
    }
}
NGINX

nginx`,
		Env: map[string]string{
			"NODE_ENV": "development",
		},
	},
	"go": {
		Name:        "go",
		DisplayName: "Go",
		Description: "Go application",
		Image:       "docker.io/library/golang:1.22-alpine",
		Webroot:     ".",
		Services:    []string{},
		InstallCmd:  `apk add --no-cache nginx git`,
		Env: map[string]string{
			"CGO_ENABLED": "0",
		},
	},
	"python": {
		Name:        "python",
		DisplayName: "Python",
		Description: "Python application",
		Image:       "docker.io/library/python:3.12-slim",
		Webroot:     ".",
		Services:    []string{"redis"},
		InstallCmd: `apt-get update -qq && apt-get install -y -qq nginx curl

# Configure nginx
cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80;
    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
    }
}
NGINX

systemctl start nginx 2>/dev/null || service nginx start`,
		Env: map[string]string{
			"PYTHONUNBUFFERED": "1",
		},
	},
	"blank": {
		Name:        "blank",
		DisplayName: "Blank",
		Description: "Empty container",
		Image:       "docker.io/library/ubuntu:24.04",
		Webroot:     ".",
		Services:    []string{},
		InstallCmd:  `apt-get update -qq && apt-get install -y -qq nginx curl`,
	},
}

// GetRecipe returns a recipe definition by name
func GetRecipe(name string) (RecipeDefinition, bool) {
	recipe, ok := RecipeLibrary[name]
	return recipe, ok
}

// ListRecipes returns all available recipes
func ListRecipes() []string {
	var names []string
	for name := range RecipeLibrary {
		names = append(names, name)
	}
	return names
}
