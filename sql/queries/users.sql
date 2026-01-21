-- name: AddUser :one
insert into
  users (id, created_at, updated_at, email)
values (
  (select gen_random_uuid()), 
  (select now()), 
  (select now()), 
  $1
)
returning *;
