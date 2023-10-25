DROP TABLE IF EXISTS [nodes];

CREATE TABLE [nodes] (
  id TEXT NOT NULL PRIMARY KEY,
  projectID TEXT NOT NULL,
  clusterName TEXT NOT NULL,
  clusterLocation TEXT NOT NULL,
  nodeName TEXT NOT NULL,
  machineType TEXT,
  preemptible BOOLEAN,
  region TEXT,
  cpuSize INTEGER,
  cpuSKU TEXT,
  memSize INTEGER,
  memSKU TEXT,
  lastUpdate DATETIME NOT NULL
);