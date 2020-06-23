create user ora_exporter identified by PASSWORD;
grant create session to ora_exporter;
grant select any dictionary to ora_exporter;
grant select_catalog_role to ora_exporter;

