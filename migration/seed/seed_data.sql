-- Seed users
INSERT INTO users (id, name, email)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Demo User', 'demo@example.com')
ON CONFLICT (id) DO NOTHING;

-- Seed collections
INSERT INTO collections (id, user_id, name, description, note_count)
VALUES
  ('f8d1b81f-1c42-49b3-9274-46997ca27688', '11111111-1111-1111-1111-111111111111', 'Gateway API', 'Edge proxy and auth pipeline', 0),
  ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Mobile UI Kit', 'Shared components and utilities', 0)
ON CONFLICT (id) DO NOTHING;

-- Seed folders
INSERT INTO folders (id, collection_id, parent_folder_id, name, sort_order)
VALUES
  ('folder-root-gateway', 'f8d1b81f-1c42-49b3-9274-46997ca27688', NULL, 'src', 0),
  ('folder-auth-gateway', 'f8d1b81f-1c42-49b3-9274-46997ca27688', 'folder-root-gateway', 'auth', 1),
  ('folder-hooks-mobile', '22222222-2222-2222-2222-222222222222', NULL, 'hooks', 0)
ON CONFLICT (id) DO NOTHING;

-- Seed notes
INSERT INTO notes (id, collection_id, folder_id, user_id, title, code, language, note, tags)
VALUES
  (
    'note-1',
    'f8d1b81f-1c42-49b3-9274-46997ca27688',
    'folder-auth-gateway',
    '11111111-1111-1111-1111-111111111111',
    'JWT exchange handler',
    $$export async function exchangeToken(input) {
  const payload = await verify(input.accessToken);
  return issue({ sub: payload.sub, aud: "internal" });
}$$,
    'TypeScript',
    'Handles the service-to-service token exchange for short lived sessions.',
    ARRAY['auth','edge']
  ),
  (
    'note-2',
    '22222222-2222-2222-2222-222222222222',
    'folder-hooks-mobile',
    '11111111-1111-1111-1111-111111111111',
    'useTimelineSync',
    $$export const useTimelineSync = (id) => {
  const query = useQuery({ queryKey: ["timeline", id], queryFn: () => fetchTimeline(id) });
  useEffect(() => {
    const sub = socket.channel(id).on("event", query.refetch);
    return () => sub.unsubscribe();
  }, [id]);
  return query;
};$$,
    'TypeScript',
    'Keeps the playback timeline synced with collaborative edits.',
    ARRAY['hooks','sync']
  )
ON CONFLICT (id) DO NOTHING;

-- Update note_count values based on notes table
UPDATE collections c
SET note_count = sub.count
FROM (
  SELECT collection_id, COUNT(*) AS count
  FROM notes
  GROUP BY collection_id
) sub
WHERE c.id = sub.collection_id;
