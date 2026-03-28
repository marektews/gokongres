# gokongres
Rest API w golangu.  
Baza nazywa się: `kongres`


## Inicjowanie konta administracyjnego

```bash
mongoimport --db kongres --collection users --file ./admin_user.json
```
Hash hasła można wygenerować na stronie: [bcrypt-generator.com](http://bcrypt-generator.com).

## Dane startowe systemu

### Różne stałe

```bash
mongoimport --db kongres --collection const --file ./const.json
```
Jeden dokument ze strukturą konfiguracji systemu.

### Działy kongresowe

```bash
mongoimport --db kongres --collection dzialy --file ./działy.json
```

### Lista zaproszonych zborów

```bash
mongoimport --db kongres --collection zbory --file ./zbory.json
```
