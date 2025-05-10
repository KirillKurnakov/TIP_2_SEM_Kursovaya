# Проект: Reverse Proxy на Nginx для Go-приложения 
## Описание 
В этом проекте реализована настройка Nginx как Reverse Proxy для Go
приложения.  
Включены балансировка нагрузки, безопасность и настройка HTTPS. 
## Технологии
Go - Nginx - Docker - Let's Encrypt (SSL) - Fail2Ban (защита от атак)      
## Структура проекта 
.
├── docker-compose.yml
├── Dockerfile
├── data
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── fail2ban
│   └── fail2ban.ini
├── go.mod
├── go.sum
├── initdb
│   └── tipOnlineShop.sql
├── main.go
├── nginx
│   └── app.conf
├── README.md
├── tasks
│   └── tasks.go
└── wait-for-it.sh
## Запуск проекта 
1. Установить Docker и Docker Compose. 
2. Склонировать репозиторий:  
git clone <URL репозитория> 
cd project-name 
3. Запустить контейнеры:  
docker-compose up -d 
4. Проверить работу по адресу: http://localhost. 
## Безопасность 
o Ограничено число запросов (DDoS-защита) 
o Настроены ограничения по IP-адресам 
o Включен HTTPS через Let's Encrypt 
## Тестирование 
o Проверить доступность сервиса:  
http://45.90.35.111 
o Проверить балансировку нагрузки:  
for i in {1..10}; do curl -s http://45.90.35.111 | grep "Server ID"; done 
o Проверить HTTPS:  
curl -I https://example.com 
## Логи 
o Nginx access log: /var/log/nginx/access.log 
o Nginx error log: /var/log/nginx/error.log 
## Выводы 
В ходе работы были изучены и применены: 
o Настройка Reverse Proxy на Nginx 
o Балансировка нагрузки между экземплярами Go-приложения 
o Настройка HTTPS с автоматическим обновлением сертификатов 
o Реализация базовых мер безопасности 