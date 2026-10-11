-- ==========================================
-- 4. VIEWS (DATA SLICES)
-- ==========================================

-- View: vw_active_projects
-- Retrieves all non-deleted projects with 'Open' status and their client details.
CREATE OR REPLACE VIEW vw_active_projects AS
SELECT p.id AS project_id, 
       p.title, 
       p.budget, 
       c.name AS category_name, 
       cp.company_name AS client, 
       p.created_at
FROM projects p
JOIN categories c ON p.category_id = c.id
JOIN client_profiles cp ON p.client_id = cp.user_id
WHERE p.is_deleted = FALSE 
  AND p.status = 'Open';

-- View: vw_freelancer_stats
-- Aggregates total active contracts and basic profile data for each freelancer.
CREATE OR REPLACE VIEW vw_freelancer_stats AS
SELECT u.id AS freelancer_id, 
       u.username, 
       fp.hourly_rate,
       COUNT(c.id) AS total_active_contracts
FROM users u
JOIN freelancer_profiles fp ON u.id = fp.user_id
LEFT JOIN proposals pr ON fp.user_id = pr.freelancer_id
LEFT JOIN contracts c ON pr.id = c.proposal_id AND c.status = 'Active'
WHERE u.is_deleted = FALSE
GROUP BY u.id, u.username, fp.hourly_rate;

-- View: vw_project_proposals
-- Displays all proposals associated with valid (non-deleted) projects and users.
CREATE OR REPLACE VIEW vw_project_proposals AS
SELECT pr.id AS proposal_id, 
       p.title AS project_title, 
       pr.bid_amount, 
       pr.status, 
       u.username AS freelancer_name
FROM proposals pr
JOIN projects p ON pr.project_id = p.id
JOIN users u ON pr.freelancer_id = u.id
WHERE p.is_deleted = FALSE 
  AND u.is_deleted = FALSE;

-- View: vw_user_messages
-- Retrieves chat history ignoring soft-deleted messages.
CREATE OR REPLACE VIEW vw_user_messages AS
SELECT m.id AS message_id, 
       m.project_id, 
       s.username AS sender, 
       r.username AS receiver, 
       m.content, 
       m.sent_at, 
       m.is_read
FROM messages m
JOIN users s ON m.sender_id = s.id
JOIN users r ON m.receiver_id = r.id
WHERE m.is_deleted = FALSE;

-- ==========================================
-- 5. STORED PROCEDURES
-- ==========================================

-- Procedure: sp_soft_delete_project
-- Sets the is_deleted flag and logs the user who performed the deletion.
CREATE OR REPLACE PROCEDURE sp_soft_delete_project(
    p_project_id INT,
    p_user_id INT
)
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE projects
    SET is_deleted = TRUE,
        updated_by = p_user_id
    WHERE id = p_project_id 
      AND is_deleted = FALSE;
END;
$$;

-- Procedure: sp_create_proposal
-- Inserts a new proposal from a freelancer for a specific project.
CREATE OR REPLACE PROCEDURE sp_create_proposal(
    p_project_id INT,
    p_freelancer_id INT,
    p_cover_letter TEXT,
    p_bid_amount DECIMAL
)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO proposals (project_id, freelancer_id, cover_letter, bid_amount)
    VALUES (p_project_id, p_freelancer_id, p_cover_letter, p_bid_amount);
END;
$$;

-- Procedure: sp_accept_proposal
-- Updates proposal status and generates a new contract automatically.
CREATE OR REPLACE PROCEDURE sp_accept_proposal(
    p_proposal_id INT,
    p_user_id INT,
    p_start_date DATE,
    p_end_date DATE
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_bid_amount DECIMAL;
BEGIN
    SELECT bid_amount INTO v_bid_amount 
    FROM proposals 
    WHERE id = p_proposal_id;
    
    UPDATE proposals
    SET status = 'Accepted', 
        updated_by = p_user_id
    WHERE id = p_proposal_id;
    
    INSERT INTO contracts (proposal_id, start_date, end_date, total_amount, updated_by)
    VALUES (p_proposal_id, p_start_date, p_end_date, v_bid_amount, p_user_id);
END;
$$;

-- Procedure: sp_update_project_status
-- Modifies project status and records the user making the change.
CREATE OR REPLACE PROCEDURE sp_update_project_status(
    p_project_id INT,
    p_status VARCHAR,
    p_user_id INT
)
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE projects
    SET status = p_status, 
        updated_by = p_user_id
    WHERE id = p_project_id;
END;
$$;

-- Procedure: sp_send_message
-- Creates a new message record between two users within a project context.
CREATE OR REPLACE PROCEDURE sp_send_message(
    p_project_id INT,
    p_sender_id INT,
    p_receiver_id INT,
    p_content TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO messages (project_id, sender_id, receiver_id, content)
    VALUES (p_project_id, p_sender_id, p_receiver_id, p_content);
END;
$$;