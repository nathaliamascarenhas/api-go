-- Gera 100 milhões de usuários de teste na tabela usuarios.
-- Insere em lotes de 1 milhão, com COMMIT por lote, para não estourar memória/WAL.
-- Os e-mails usam o prefixo "seed", então não colidem com dados existentes
-- e podem ser removidos depois com:
--   DELETE FROM usuarios WHERE email LIKE 'seed%@exemplo.com';

CREATE OR REPLACE PROCEDURE seed_usuarios(total bigint, lote bigint)
LANGUAGE plpgsql
AS $$
DECLARE
    inicio bigint := 1;
    fim    bigint;
BEGIN
    WHILE inicio <= total LOOP
        fim := LEAST(inicio + lote - 1, total);


        INSERT INTO usuarios (nome, email, telefone)
        SELECT
            (ARRAY['Ana','Bruno','Carla','Diego','Elisa','Fernando','Gabriela','Hugo','Irene','João'])[1 + floor(random() * 10)::int]
              || ' ' ||
            (ARRAY['Silva','Souza','Costa','Oliveira','Pereira','Rodrigues','Almeida','Lima','Gomes','Ribeiro'])[1 + floor(random() * 10)::int],
            'seed' || g || '@exemplo.com',
            '119' || lpad(floor(random() * 100000000)::bigint::text, 8, '0')
        FROM generate_series(inicio, fim) AS g;


        COMMIT;
        RAISE NOTICE 'Inseridos % de % registros', fim, total;
        inicio := fim + 1;
    END LOOP;
END;
$$;

CALL seed_usuarios(100000000, 1000000);

DROP PROCEDURE seed_usuarios(bigint, bigint);
