alter table repairs
    add column if not exists front_photo_path text,
    add column if not exists back_photo_path text;
