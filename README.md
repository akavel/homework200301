**Uwaga ogólna:** w projekcie pozostawiłem komentarze "FIXME" i "TODO" na oznaczenie różnych aspektów rozwojowych dla projektu, które można podjąć zależnie od dalszych potrzeb i wymagań. 

Uwagi szczegółowe nawiązujące do niektórych punktów zadania:

**Ad 2.:** przyjąłem angielskie nazewnictwo w API

**Ad 6.:** wybrałem PostgreSQL

**Ad 7.:** operacje REST udostępniane przez serwis:

- `GET localhost:8080/v1/user` &mdash; "Wylistowanie wszystkich użytkowników z bazy danych [...]"
  - `?deleted=*` &mdash; "lista wszystkich użytkowników" 
  - **domyślnie** (lub `?deleted=false&technology=*`) &mdash; "lista aktywnych użytkowników"
  - `?deleted=true` &mdash; "lista usuniętych użytkowników"
  - `?technology=go` &mdash; przykładowe "filtrowanie po polu technologia"
- `GET localhost:8080/v1/user/$EMAIL` &mdash; "Pobranie danych dowolnego użytkownika po podaniu jego identyfikatora"
- `POST localhost:8080/v1/user` &mdash; "Stworzenie nowego użytkownika"
- `PUT localhost:8080/v1/user/$EMAIL` &mdash; "Edycja danych użytkownika"\
- `DELETE localhost:8080/v1/user/$EMAIL` &mdash; "Usunięcie użytkownika (soft delete)"

**Ad 9.:** plik tekstowy `requests.log` tworzony jest w wolumenie dockera o nazwie: `users_logs`

**Ad 10.:** `docker-compose up -d --build`
