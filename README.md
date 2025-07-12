# Bookmark-RESTful

RESTful API for the Bookmark saving application

To run in the terminal:
POST :
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{"name":"example", "email":"example@example.com"}'

GET:
curl http://localhost:8080/users/1/bookmarks
