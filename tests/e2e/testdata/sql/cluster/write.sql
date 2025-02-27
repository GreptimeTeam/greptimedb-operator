CREATE TABLE my_table (
                          a INT PRIMARY KEY,
                          b STRING,
                          ts TIMESTAMP TIME INDEX,
)
    PARTITION ON COLUMNS (a) (
  a < 1000,
  a >= 1000 AND a < 2000,
  a >= 2000
);

INSERT INTO my_table VALUES
                         (100, 'a', 1),
                         (200, 'b', 2),
                         (1100, 'c', 3),
                         (1200, 'd', 4),
                         (2000, 'e', 5),
                         (2100, 'f', 6),
                         (2200, 'g', 7),
                         (2400, 'h', 8);

DELETE FROM my_table WHERE a < 150;
