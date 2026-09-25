CREATE INDEX idx_conversations_buyer
    ON conversations(buyer_id);

CREATE INDEX idx_conversations_farmer
    ON conversations(farmer_id);

CREATE INDEX idx_conversations_offering
    ON conversations(offering_id);

CREATE INDEX idx_messages_conversation_created
    ON messages(conversation_id, created_at);
