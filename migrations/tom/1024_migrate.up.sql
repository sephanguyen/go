update conversations c set resource_path = c.owner
where c.conversation_type = ANY('{CONVERSATION_STUDENT,CONVERSATION_PARENT}')
and (c.resource_path is null or length(c.resource_path) =0);

update conversation_students cs set resource_path=c.resource_path
from conversations c
where cs.conversation_id=c.conversation_id 
and c.conversation_type = ANY('{CONVERSATION_STUDENT,CONVERSATION_PARENT}')
and (cs.resource_path is null or length(cs.resource_path)=0);

update conversation_members cm set resource_path=c.resource_path
from conversations c
where cm.conversation_id=c.conversation_id 
and c.conversation_type = ANY('{CONVERSATION_STUDENT,CONVERSATION_PARENT}')
and (cm.resource_path is null or length(cm.resource_path)=0);

update messages m set resource_path=c.resource_path
from conversations c
where m.conversation_id=c.conversation_id 
and c.conversation_type = ANY('{CONVERSATION_STUDENT,CONVERSATION_PARENT}')
and (m.resource_path is null or length(m.resource_path)=0);
