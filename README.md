# Bookmark-RESTful

RESTful API for the Bookmark saving application

To run in the terminal:
POST :
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{"name":"example", "email":"example@example.com"}'

GET:
curl http://localhost:8080/users/1/bookmarks

Bookmark-RESTful ➤ curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name": "Test User", "email": "test@example.com"}'
{"id":17,"name":"Test User","email":"test@example.com","api_key":"a5d141ca19d7a8c96ee3d30fe4feb13c4c7ceb4b22f6cbb6401a4d060ad4b610"}

Bookmark-RESTful ➤ curl -X GET http://localhost:8080/users/1/bookmarks
[{"id":1,"user_id":1,"title":"OpenAI Homepage","url":"https://openai.com","created_at":"2025-07-07T20:42:45Z"},{"id":2,"user_id":1,"title":"Google","url":"https://google.com","created_at":"2025-07-07T20:43:32Z"}]
