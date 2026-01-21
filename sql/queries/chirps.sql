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
