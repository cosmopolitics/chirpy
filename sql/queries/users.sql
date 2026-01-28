-- name: Reset :exec
delete from users where id = id;
--

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

-- name: GetUserByEmail :one
select * from
  users
where 
  email = $1;
--

-- name: SubcribeUser :one
update  
  users
set
  is_chirpy_red = true
where 
  id = $1
returning 
  true;
--
