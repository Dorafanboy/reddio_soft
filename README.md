1. Установить go с официального сайта Go https://go.dev/dl/
2. Скачать проект, можно через ```git clone https://github.com/Dorafanboy/reddio_soft.git```
3. Установить зависимости, ```go mod tidy```
4. Заполнить данные в папке data.
Приватные ключи заполнить в private_keys, можно указывать с '0x' или без, 1 строка 1 приватный ключ, прокси указывать в формате username:login@ip:port, в register_codes указывать коды регистрации, 1 строка 1 код, они выбираются рандомно,
в twitter_data указать данные в формате CT0:AUTH_TOKEN 1 строка 1 данные твитера, в config.yaml в папке data указать задержки.
5. ```make run``` чтобы запустить скрипт