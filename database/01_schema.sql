-- ==========================================
-- 1. TRIGGERS & FUNCTIONS
-- ==========================================

-- Function to automatically update the 'updated_at' timestamp.
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- 2. TABLES CREATION
-- ==========================================

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    role_id INT REFERENCES roles(id) ON DELETE RESTRICT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP,
    updated_by INT REFERENCES users(id) ON DELETE SET NULL
);

-- Attach the audit trigger to the users table
CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE freelancer_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    bio TEXT,
    hourly_rate DECIMAL(10, 2)
);

CREATE TABLE client_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(100),
    contact_person VARCHAR(100),
    billing_address TEXT
);

CREATE TABLE skills (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE freelancer_skills (
    freelancer_id INT REFERENCES freelancer_profiles(user_id) ON DELETE CASCADE,
    skill_id INT REFERENCES skills(id) ON DELETE CASCADE,
    proficiency_level VARCHAR(20),
    PRIMARY KEY (freelancer_id, skill_id)
);

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    client_id INT REFERENCES client_profiles(user_id) ON DELETE RESTRICT,
    category_id INT REFERENCES categories(id) ON DELETE RESTRICT,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    budget DECIMAL(12, 2),
    status VARCHAR(20) DEFAULT 'Open',
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP,
    updated_by INT REFERENCES users(id) ON DELETE SET NULL
);

CREATE TRIGGER set_timestamp_projects
BEFORE UPDATE ON projects
FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE project_skills (
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
    skill_id INT REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, skill_id)
);

CREATE TABLE proposals (
    id SERIAL PRIMARY KEY,
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
    freelancer_id INT REFERENCES freelancer_profiles(user_id) ON DELETE CASCADE,
    cover_letter TEXT,
    bid_amount DECIMAL(12, 2),
    status VARCHAR(20) DEFAULT 'Pending',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP,
    updated_by INT REFERENCES users(id) ON DELETE SET NULL
);

CREATE TRIGGER set_timestamp_proposals
BEFORE UPDATE ON proposals
FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE contracts (
    id SERIAL PRIMARY KEY,
    proposal_id INT UNIQUE REFERENCES proposals(id) ON DELETE RESTRICT,
    start_date DATE,
    end_date DATE,
    total_amount DECIMAL(12, 2),
    status VARCHAR(20) DEFAULT 'Active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP,
    updated_by INT REFERENCES users(id) ON DELETE SET NULL
);

CREATE TRIGGER set_timestamp_contracts
BEFORE UPDATE ON contracts
FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE milestones (
    id SERIAL PRIMARY KEY,
    contract_id INT REFERENCES contracts(id) ON DELETE CASCADE,
    description VARCHAR(255),
    amount DECIMAL(12, 2),
    due_date DATE,
    is_completed BOOLEAN DEFAULT FALSE
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    contract_id INT REFERENCES contracts(id) ON DELETE RESTRICT,
    milestone_id INT REFERENCES milestones(id) ON DELETE SET NULL,
    amount DECIMAL(12, 2),
    transaction_date TIMESTAMP DEFAULT NOW(),
    payment_method VARCHAR(50)
);

CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    contract_id INT UNIQUE REFERENCES contracts(id) ON DELETE CASCADE,
    reviewer_id INT REFERENCES users(id) ON DELETE CASCADE,
    reviewee_id INT REFERENCES users(id) ON DELETE CASCADE,
    rating INT CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    project_id INT REFERENCES projects(id) ON DELETE CASCADE,
    sender_id INT REFERENCES users(id) ON DELETE CASCADE,
    receiver_id INT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    sent_at TIMESTAMP DEFAULT NOW(),
    is_read BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE message_attachments (
    id SERIAL PRIMARY KEY,
    message_id INT REFERENCES messages(id) ON DELETE CASCADE,
    file_url VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

-- ==========================================
-- 3. INDEXES (Two different types)
-- ==========================================

-- B-Tree Indexes for fast foreign key lookups and sorting.
CREATE INDEX idx_projects_client_id ON projects(client_id);
CREATE INDEX idx_proposals_project_id ON proposals(project_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);

-- GIN Index (Generalized Inverted Index) for full-text search.
-- Converts project description into tsvector for fast keyword searching.
CREATE INDEX idx_projects_description_gin 
ON projects USING GIN (to_tsvector('english', description));