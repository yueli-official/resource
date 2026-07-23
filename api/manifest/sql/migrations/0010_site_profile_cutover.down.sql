DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM site_profile_state
        WHERE document #>> '{branding,logo,kind}' = 'asset'
    ) THEN
        RAISE EXCEPTION 'cannot downgrade Resource Site Profile with an Asset-backed logo';
    END IF;
END $$;

WITH profile AS (
    SELECT document
    FROM site_profile_state
    WHERE id = 1
),
mapped AS (
    SELECT
        jsonb_build_object(
            'siteName', document #>> '{identity,name}',
            'tagline', COALESCE(document #>> '{identity,tagline}', ''),
            'logoIcon', COALESCE(document #>> '{branding,logo,ref}', 'i-tabler-package'),
            'announcement', COALESCE(document #>> '{announcement,text}', ''),
            'announcementEnabled', COALESCE((document #>> '{announcement,enabled}')::boolean, false),
            'supportEmail', COALESCE(
                jsonb_path_query_first(document, '$.support.contacts[*] ? (@.kind == "email") .value') #>> '{}',
                ''
            )
        ) AS site,
        jsonb_build_object(
            'tagline', COALESCE(document #>> '{footer,tagline}', ''),
            'copyright', COALESCE(document #>> '{footer,copyright}', ''),
            'compliance', jsonb_build_object(
                'icpRecord', COALESCE(jsonb_path_query_first(document, '$.footer.compliance.records[*] ? (@.kind == "icp") .number') #>> '{}', ''),
                'icpUrl', COALESCE(jsonb_path_query_first(document, '$.footer.compliance.records[*] ? (@.kind == "icp") .url') #>> '{}', ''),
                'policeRecord', COALESCE(jsonb_path_query_first(document, '$.footer.compliance.records[*] ? (@.kind == "police") .number') #>> '{}', ''),
                'policeUrl', COALESCE(jsonb_path_query_first(document, '$.footer.compliance.records[*] ? (@.kind == "police") .url') #>> '{}', ''),
                'extraText', COALESCE(document #>> '{footer,compliance,extraText}', '')
            ),
            'linkGroups', COALESCE((
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'title', item ->> 'title',
                        'links', COALESCE((
                            SELECT jsonb_agg(jsonb_build_object(
                                'label', link ->> 'label',
                                'to', link ->> 'href',
                                'icon', COALESCE(link ->> 'icon', '')
                            ))
                            FROM jsonb_array_elements(item -> 'links') AS link
                        ), '[]'::jsonb)
                    )
                )
                FROM jsonb_array_elements(document #> '{footer,linkGroups}') AS item
            ), '[]'::jsonb),
            'socialLinks', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                    'label', COALESCE(item ->> 'label', item ->> 'platform'),
                    'to', item ->> 'url',
                    'icon', COALESCE(item ->> 'icon', '')
                ))
                FROM jsonb_array_elements(document #> '{footer,social}') AS item
            ), '[]'::jsonb)
        ) AS footer
    FROM profile
)
UPDATE resource_site_settings
SET payload = payload || jsonb_build_object('site', mapped.site, 'footer', mapped.footer)
FROM mapped
WHERE key = 'site';
