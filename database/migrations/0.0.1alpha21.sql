ALTER TABLE materials
ADD COLUMN filename TEXT not null DEFAULT 'infomark_zip_or_pdf';
-- unfortunately, we do not the extension from previous uploads
-- but they be only a zip or pdf file
UPDATE materials set filename=name || '.zip_or_pdf';