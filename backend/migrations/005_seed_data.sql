INSERT INTO users (id, name, email) VALUES
    ('a1111111-1111-1111-1111-111111111111', 'Alice', 'alice@example.com'),
    ('b2222222-2222-2222-2222-222222222222', 'Bob', 'bob@example.com'),
    ('c3333333-3333-3333-3333-333333333333', 'Charlie', 'charlie@example.com'),
    ('d4444444-4444-4444-4444-444444444444', 'Dave', 'dave@example.com')
ON CONFLICT (id) DO NOTHING;

INSERT INTO events (id, title, description, capacity, booked_count, event_date, location) VALUES
    ('e1111111-1111-1111-1111-111111111111', 'Go Workshop 2025',
     'Advanced concurrency patterns in Go', 3, 0,
     '2025-08-15 10:00:00+05:30', 'Noida'),
    ('e2222222-2222-2222-2222-222222222222', 'React Meetup',
     'Hooks deep dive and performance', 20, 0,
     '2025-08-20 18:00:00+05:30', 'Delhi'),
    ('e3333333-3333-3333-3333-333333333333', 'System Design Talk',
     'Designing for scale - lessons from production', 2, 0,
     '2025-09-01 14:00:00+05:30', 'Bangalore'),
    ('e4444444-4444-4444-4444-444444444444', 'AI/ML Hackathon',
     'Build something cool with LLMs in 24 hours', 50, 0,
     '2025-09-10 09:00:00+05:30', 'Noida'),
    ('e5555555-5555-5555-5555-555555555555', 'DevOps Bootcamp',
     'K8s, Docker, CI/CD hands-on', 10, 0,
     '2025-09-15 10:00:00+05:30', 'Gurugram')
ON CONFLICT (id) DO NOTHING;
