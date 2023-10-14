DROP FUNCTION IF EXISTS public.calculate_learning_time();

CREATE or replace FUNCTION public.calculate_learning_time(_student_id text)
    RETURNS TABLE
            (
                learning_time_by_minutes    int,
                student_id                  text,
                sessions                    text,
                date                        date,
                day                         timestamptz,
                assignment_duration         int,
                submit_learning_material_id text
            )
    LANGUAGE sql
    STABLE
AS
$$
with data as (select created_at                              as date
                   , to_date(created_at::text, 'YYYY-MM-DD') as date_normalize
                   , payload ->> 'session_id'                as session_id
                   , payload ->> 'event'                     as event
                   , student_id
              from student_event_logs
              where 1 = 1
                and payload ->> 'session_id' is not null
                and payload ->> 'event' is not null
                and event_type = 'learning_objective'
                and student_id = _student_id
              order by created_at)
   , pause_resume_normalize as (select student_id,
                                       session_id,
                                       event,
                                       event_at,
                                       next_event,
                                       next_event_at,
                                       date_normalize
                                from (select student_id,
                                             session_id,
                                             date_normalize
                                              ,
                                             lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                              ,
                                             lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                              ,
                                             lead(session_id) over (order by session_id, date)     as next_session_id,
                                             lag(session_id) over (order by session_id, date)      as prev_session_id,
                                             event,
                                             date                                                  as event_at,
                                             lag(event) over (order by session_id, date)           as prev_event,
                                             lead(event) over (order by session_id, date)          as next_event,
                                             lead(date) over (order by session_id, date)           as next_event_at
                                      from (select session_id
                                                 , date
                                                 , date_normalize
                                                 , event
                                                 , lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                 , lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                 , lead(session_id) over (order by session_id, date)     as next_session_id
                                                 , lag(session_id) over (order by session_id, date)      as prev_session_id
                                                 , lead(event) over (order by session_id, date)          as next_event
                                                 , lag(event) over (order by session_id, date)           as prev_event
                                                 , student_id
                                            from data
                                            where event = 'paused'
                                               or event = 'resumed'
                                               or event = 'started'
                                               or event = 'exited') pause_resume
                                      where ((event = 'paused' and
                                              (prev_event is null or
                                               (not (prev_event = 'paused' or prev_event = 'exited') and
                                                session_id = prev_session_id) or
                                               (prev_event = 'paused' and session_id != prev_session_id)) and
                                              (next_event is not null and
                                               (not next_event = 'started' or prev_event = 'started')) and
                                              ((date_normalize = next_date_normalize or
                                                date_normalize = prev_date_normalize) and
                                               (session_id = next_session_id or session_id = prev_session_id))) or
                                             (event = 'resumed' and next_event is null and
                                              (session_id = next_session_id or session_id = prev_session_id)) or
                                             (event = 'resumed' and (not next_event = 'resumed') and
                                              (prev_event is not null and not prev_event = 'resumed') and
                                              (session_id = next_session_id or session_id = prev_session_id)))
                                      order by student_id, session_id, date) pause_resume_filter
                                where event = 'paused'
                                  and (next_event = 'resumed' and next_session_id = session_id)
                                  and (session_id = next_session_id or session_id = prev_session_id))
   , start_complete_normalize as (select student_id,
                                         session_id,
                                         event,
                                         event_at,
                                         next_event,
                                         next_event_at,
                                         date_normalize
                                  from (select session_id,
                                               date_normalize
                                                ,
                                               lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                ,
                                               lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                ,
                                               lead(session_id) over (order by session_id, date)     as next_session_id,
                                               lag(session_id) over (order by session_id, date)      as prev_session_id,
                                               event,
                                               date                                                  as event_at,
                                               lag(event) over (order by session_id, date)           as prev_event,
                                               lead(event) over (order by session_id, date)          as next_event,
                                               lead(date) over (order by session_id, date)           as next_event_at,
                                               student_id
                                        from (select session_id
                                                   , date
                                                   , date_normalize
                                                   , event
                                                   , lead(date_normalize) over (order by session_id, date) as next_date_normalize
                                                   , lag(date_normalize) over (order by session_id, date)  as prev_date_normalize
                                                   , lead(session_id) over (order by session_id, date)     as next_session_id
                                                   , lag(session_id) over (order by session_id, date)      as prev_session_id
                                                   , lead(event) over (order by session_id, date)          as next_event
                                                   , lag(event) over (order by session_id, date)           as prev_event
                                                   , student_id
                                              from data
                                              where event = 'started'
                                                 or event = 'completed') start_complete
                                        where (event = 'started' and (prev_event is null or prev_event != 'started' or
                                                                      (prev_event = 'started' and session_id != prev_session_id))
                                            and ((date_normalize = next_date_normalize or
                                                  date_normalize = prev_date_normalize) and
                                                 (session_id = next_session_id or session_id = prev_session_id)))
                                           or (event = 'completed' and next_event is null and
                                               ((date_normalize = next_date_normalize or
                                                 date_normalize = prev_date_normalize) and
                                                (session_id = next_session_id or session_id = prev_session_id)))
                                           or (event = 'completed' and not next_event = 'completed' and
                                               prev_event is not null and ((date_normalize = next_date_normalize or
                                                                            date_normalize = prev_date_normalize) and
                                                                           (session_id = next_session_id or session_id = prev_session_id)))) start_complete_filter
                                  where event = 'started'
                                    and (next_event = 'completed' and next_session_id = session_id)
                                    and (session_id = next_session_id or session_id = prev_session_id))
   , total_sc_by_date as (select sum(time_by_sec)      as time_by_sec,
                                 student_id,
                                 date_normalize,
                                 array_agg(session_id) as sessions
                          from (select (coalesce(EXTRACT(epoch FROM (sum(next_event_at - event_at))), 0)) as time_by_sec,
                                       student_id,
                                       session_id,
                                       date_normalize
                                from start_complete_normalize
                                group by student_id, session_id, date_normalize) total_sc_by_session
                          group by student_id, date_normalize)
   , total_pr_by_date as (select sum(time_by_sec)      as time_by_sec,
                                 student_id,
                                 date_normalize,
                                 array_agg(session_id) as sessions
                          from (select (coalesce(EXTRACT(epoch FROM (sum(next_event_at - event_at))), 0)) as time_by_sec,
                                       student_id,
                                       session_id,
                                       date_normalize
                                from pause_resume_normalize
                                group by student_id, session_id, date_normalize) total_pr_by_session
                          group by student_id, date_normalize)
   , learning_time as (select distinct ((a.time_by_sec - coalesce(b.time_by_sec, 0)) / 60)::int as minute,
                                       a.student_id,
                                       (select array_agg(sessions)::text
                                        from (select distinct unnest(aa.sessions || b.sessions) as sessions
                                              from total_sc_by_date aa
                                                       left join total_pr_by_date b using (student_id, date_normalize)
                                              where a.student_id = aa.student_id
                                                and a.date_normalize = aa.date_normalize) subq) as sessions,
                                       a.date_normalize                                         as date,
                                       a.date_normalize::timestamp AT TIME ZONE 'UTC'           as day
                       from total_sc_by_date a
                                left join total_pr_by_date b using (student_id, date_normalize)
                       group by a.student_id, a.date_normalize, a.time_by_sec, b.time_by_sec)
select (learning_time.minute + coalesce(sum(student_submissions.duration),0) / 60)::int,
       learning_time.student_id,
       learning_time.sessions,
       learning_time.date,
       learning_time.day,
       (sum(student_submissions.duration) / 60)::int as assignment_duration,
       (array_agg(learning_material_id))::text             as learning_material_ids
from learning_time
         left join student_submissions on learning_time.student_id = student_submissions.student_id and learning_time.date = to_date(student_submissions.created_at::text, 'YYYY-MM-DD')
group by learning_time.minute, learning_time.student_id, learning_time.sessions, learning_time.date, learning_time.day
$$;
