-- name: AddChirp :one
insert into 
  chirps (id, created_at, updated_at, body, uid)
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
