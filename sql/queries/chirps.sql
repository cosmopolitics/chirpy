-- name: AddChirp :one
insert into 
  chirps (id, created_at, updated_at, body, user_id)
values 
  ((select gen_random_uuid()), 
  (select now()),
  (select now()),
  $1, 
  $2)
returning *;
--

-- name: GetAllChirps :many 
select 
  * 
from 
  chirps
order by
  created_at;
--

-- name: GetChirpById :one
select 
  *
from 
  chirps
where 
  id = $1;
--

-- name: DeleteChirp :exec
delete from 
  chirps
where 
  id = $1;
--

-- name: GetUsersChirps :many
select 
  *
from
  chirps
where 
  user_id = $1
order by
  created_at;
--


