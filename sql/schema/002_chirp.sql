-- +goose Up
create table chirps (
  id uuid primary key,
  created_at timestamp not null,
  updated_at timestamp not null,
  body text not null, 
  uid uuid not null,
    foreign key (uid)
    references users(id)
    on delete cascade
);

-- +goose Down
drop table chirps;
