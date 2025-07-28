#!/bin/bash
# scripts/neo4j-setup.sh - Script to initialize Neo4j with organizational data

set -e

echo "Setting up Neo4j for Cube Castle organizational graph..."

# Neo4j connection details
NEO4J_URI="bolt://localhost:7687"
NEO4J_USER="neo4j"
NEO4J_PASSWORD="password123"

# Wait for Neo4j to be ready
echo "Waiting for Neo4j to be ready..."
until cypher-shell -a $NEO4J_URI -u $NEO4J_USER -p $NEO4J_PASSWORD "RETURN 1" > /dev/null 2>&1; do
  echo "Waiting for Neo4j..."
  sleep 5
done

echo "Neo4j is ready! Creating schema and constraints..."

# Create constraints and indexes
cypher-shell -a $NEO4J_URI -u $NEO4J_USER -p $NEO4J_PASSWORD << 'EOF'
// Employee constraints
CREATE CONSTRAINT employee_id_unique IF NOT EXISTS FOR (e:Employee) REQUIRE e.employee_id IS UNIQUE;
CREATE CONSTRAINT employee_uuid_unique IF NOT EXISTS FOR (e:Employee) REQUIRE e.id IS UNIQUE;

// Position constraints
CREATE CONSTRAINT position_id_unique IF NOT EXISTS FOR (p:Position) REQUIRE p.id IS UNIQUE;

// Department constraints  
CREATE CONSTRAINT department_name_unique IF NOT EXISTS FOR (d:Department) REQUIRE d.name IS UNIQUE;

// Team constraints
CREATE CONSTRAINT team_id_unique IF NOT EXISTS FOR (t:Team) REQUIRE t.id IS UNIQUE;

// Employee indexes
CREATE INDEX employee_legal_name IF NOT EXISTS FOR (e:Employee) ON (e.legal_name);
CREATE INDEX employee_email IF NOT EXISTS FOR (e:Employee) ON (e.email);
CREATE INDEX employee_status IF NOT EXISTS FOR (e:Employee) ON (e.status);
CREATE INDEX employee_hire_date IF NOT EXISTS FOR (e:Employee) ON (e.hire_date);

// Position indexes
CREATE INDEX position_title IF NOT EXISTS FOR (p:Position) ON (p.position_title);
CREATE INDEX position_department IF NOT EXISTS FOR (p:Position) ON (p.department);
CREATE INDEX position_job_level IF NOT EXISTS FOR (p:Position) ON (p.job_level);
CREATE INDEX position_effective_date IF NOT EXISTS FOR (p:Position) ON (p.effective_date);
CREATE INDEX position_location IF NOT EXISTS FOR (p:Position) ON (p.location);

// Department indexes
CREATE INDEX department_parent IF NOT EXISTS FOR (d:Department) ON (d.parent_id);
CREATE INDEX department_manager IF NOT EXISTS FOR (d:Department) ON (d.manager_id);

// Relationship indexes
CREATE INDEX reports_to_date IF NOT EXISTS FOR ()-[r:REPORTS_TO]-() ON (r.effective_date);
CREATE INDEX holds_position_date IF NOT EXISTS FOR ()-[r:HOLDS_POSITION]-() ON (r.effective_date);
CREATE INDEX belongs_to_date IF NOT EXISTS FOR ()-[r:BELONGS_TO]-() ON (r.effective_date);

RETURN "Schema created successfully" as result;
EOF

echo "Creating sample organizational data..."

# Create sample data for testing
cypher-shell -a $NEO4J_URI -u $NEO4J_USER -p $NEO4J_PASSWORD << 'EOF'
// Create departments
MERGE (tech:Department {name: "技术部"})
SET tech.id = "dept-tech",
    tech.created_at = datetime(),
    tech.description = "Technology and Engineering Department";

MERGE (product:Department {name: "产品部"})
SET product.id = "dept-product",
    product.created_at = datetime(),
    product.description = "Product Management Department";

MERGE (sales:Department {name: "销售部"})
SET sales.id = "dept-sales",
    sales.created_at = datetime(),
    sales.description = "Sales and Business Development Department";

MERGE (hr:Department {name: "人力资源部"})
SET hr.id = "dept-hr",
    hr.created_at = datetime(),
    hr.description = "Human Resources Department";

// Create sample employees
MERGE (emp1:Employee {employee_id: "EMP001"})
SET emp1.id = "uuid-emp-001",
    emp1.legal_name = "张三",
    emp1.email = "zhang.san@company.com",
    emp1.status = "ACTIVE",
    emp1.hire_date = datetime("2020-01-15T00:00:00Z"),
    emp1.created_at = datetime();

MERGE (emp2:Employee {employee_id: "EMP002"})
SET emp2.id = "uuid-emp-002",
    emp2.legal_name = "李四",
    emp2.email = "li.si@company.com",
    emp2.status = "ACTIVE",
    emp2.hire_date = datetime("2021-03-20T00:00:00Z"),
    emp2.created_at = datetime();

MERGE (emp3:Employee {employee_id: "EMP003"})
SET emp3.id = "uuid-emp-003",
    emp3.legal_name = "王五",
    emp3.email = "wang.wu@company.com",
    emp3.status = "ACTIVE",
    emp3.hire_date = datetime("2022-06-10T00:00:00Z"),
    emp3.created_at = datetime();

MERGE (emp4:Employee {employee_id: "EMP004"})
SET emp4.id = "uuid-emp-004",
    emp4.legal_name = "赵六",
    emp4.email = "zhao.liu@company.com",
    emp4.status = "ACTIVE",
    emp4.hire_date = datetime("2023-02-15T00:00:00Z"),
    emp4.created_at = datetime();

// Create positions
MERGE (pos1:Position {id: "pos-001"})
SET pos1.position_title = "技术总监",
    pos1.department = "技术部",
    pos1.job_level = "DIRECTOR",
    pos1.location = "北京",
    pos1.effective_date = datetime("2020-01-15T00:00:00Z"),
    pos1.created_at = datetime();

MERGE (pos2:Position {id: "pos-002"})
SET pos2.position_title = "高级软件工程师",
    pos2.department = "技术部",
    pos2.job_level = "SENIOR",
    pos2.location = "北京",
    pos2.effective_date = datetime("2021-03-20T00:00:00Z"),
    pos2.created_at = datetime();

MERGE (pos3:Position {id: "pos-003"})
SET pos3.position_title = "前端工程师",
    pos3.department = "技术部",
    pos3.job_level = "INTERMEDIATE",
    pos3.location = "上海",
    pos3.effective_date = datetime("2022-06-10T00:00:00Z"),
    pos3.created_at = datetime();

MERGE (pos4:Position {id: "pos-004"})
SET pos4.position_title = "产品经理",
    pos4.department = "产品部",
    pos4.job_level = "MANAGER",
    pos4.location = "北京",
    pos4.effective_date = datetime("2023-02-15T00:00:00Z"),
    pos4.created_at = datetime();

// Create relationships
MATCH (emp1:Employee {employee_id: "EMP001"}), (pos1:Position {id: "pos-001"})
MERGE (emp1)-[r1:HOLDS_POSITION]->(pos1)
SET r1.effective_date = datetime("2020-01-15T00:00:00Z"),
    r1.created_at = datetime();

MATCH (emp2:Employee {employee_id: "EMP002"}), (pos2:Position {id: "pos-002"})
MERGE (emp2)-[r2:HOLDS_POSITION]->(pos2)
SET r2.effective_date = datetime("2021-03-20T00:00:00Z"),
    r2.created_at = datetime();

MATCH (emp3:Employee {employee_id: "EMP003"}), (pos3:Position {id: "pos-003"})
MERGE (emp3)-[r3:HOLDS_POSITION]->(pos3)
SET r3.effective_date = datetime("2022-06-10T00:00:00Z"),
    r3.created_at = datetime();

MATCH (emp4:Employee {employee_id: "EMP004"}), (pos4:Position {id: "pos-004"})
MERGE (emp4)-[r4:HOLDS_POSITION]->(pos4)
SET r4.effective_date = datetime("2023-02-15T00:00:00Z"),
    r4.created_at = datetime();

// Create department relationships
MATCH (pos1:Position {id: "pos-001"}), (tech:Department {name: "技术部"})
MERGE (pos1)-[rd1:BELONGS_TO]->(tech);

MATCH (pos2:Position {id: "pos-002"}), (tech:Department {name: "技术部"})
MERGE (pos2)-[rd2:BELONGS_TO]->(tech);

MATCH (pos3:Position {id: "pos-003"}), (tech:Department {name: "技术部"})
MERGE (pos3)-[rd3:BELONGS_TO]->(tech);

MATCH (pos4:Position {id: "pos-004"}), (product:Department {name: "产品部"})
MERGE (pos4)-[rd4:BELONGS_TO]->(product);

// Create reporting relationships
MATCH (emp2:Employee {employee_id: "EMP002"}), (emp1:Employee {employee_id: "EMP001"})
MERGE (emp2)-[rep1:REPORTS_TO]->(emp1)
SET rep1.effective_date = datetime("2021-03-20T00:00:00Z"),
    rep1.created_at = datetime();

MATCH (emp3:Employee {employee_id: "EMP003"}), (emp1:Employee {employee_id: "EMP001"})
MERGE (emp3)-[rep2:REPORTS_TO]->(emp1)
SET rep2.effective_date = datetime("2022-06-10T00:00:00Z"),
    rep2.created_at = datetime();

RETURN "Sample data created successfully" as result;
EOF

echo "Creating useful organizational queries..."

# Create some useful stored procedures for common queries
cypher-shell -a $NEO4J_URI -u $NEO4J_USER -p $NEO4J_PASSWORD << 'EOF'
// Verify our data structure
MATCH (e:Employee)-[hp:HOLDS_POSITION]->(p:Position)-[bt:BELONGS_TO]->(d:Department)
OPTIONAL MATCH (e)-[rt:REPORTS_TO]->(manager:Employee)
RETURN e.legal_name as employee,
       p.position_title as position,
       d.name as department,
       manager.legal_name as manager,
       e.hire_date as hire_date
ORDER BY d.name, p.job_level DESC;
EOF

echo "Neo4j setup completed successfully!"
echo ""
echo "Neo4j Browser: http://localhost:7474"
echo "Username: neo4j"
echo "Password: password123"
echo ""
echo "Test the setup with:"
echo "MATCH (e:Employee) RETURN count(e) as employee_count;"