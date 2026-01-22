-- name: AddUser :one
insert into
  users (id, created_at, updated_at, email, password)
values (
  (select gen_random_uuid()), 
  (select now()), 
  (select now()), 
  $1,
  $2
)
returning *;
--

-- name: Reset :exec
delete from users where id = id;
