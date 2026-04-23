INSERT INTO faculties (name) VALUES
    ('School of Information Technology'),
    ('School of Business'),
    ('School of Engineering'),
    ('School of Natural Sciences')
ON CONFLICT (name) DO NOTHING;
