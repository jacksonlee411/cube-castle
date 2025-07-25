

# **针对在WSL2 Docker中部署Temporal的根本原因分析与加固指南**

## **I. 执行摘要**

**目标：** 本报告旨在对在Windows Subsystem for Linux (WSL) 2环境下的Docker容器中部署Temporal服务及UI时遇到的顽固问题——Temporal服务持续处于“不健康”状态以及Temporal UI返回500内部服务器错误——进行深入的根本原因分析（Root Cause Analysis, RCA），并提供一套全面、可落地的解决方案与加固策略。

**核心发现：** 经过系统性排查与分析，本报告确认，用户所面临的Temporal服务“不健康”与UI 500错误并非孤立事件。它们是同一系统性不稳定问题的并发症，彼此之间存在直接的因果关联。问题的核心在于，一个不健康的后端服务必然导致依赖其数据的前端UI无法正常工作。

**已识别的根本原因：** 这种系统性不稳定性源于三个层面问题的叠加效应：

1. **Docker Compose配置的脆弱性：** Temporal官方提供的标准docker-compose.yml文件为了通用性与简洁性，并未包含针对所有依赖服务的健壮、多阶段的健康检查。这导致了服务启动过程中存在严重的服务间依赖就绪状态的竞争条件（Race Condition），特别是Temporal核心服务对其数据库（PostgreSQL）和可见性存储（Elasticsearch）的依赖。  
2. **WSL2环境的资源与性能制约：** 默认配置下的WSL2虚拟机存在显著的资源分配（尤其是内存）和I/O性能（文件系统交互）瓶颈。对于Temporal这样一个依赖数据库和搜索引擎的复杂、有状态的分布式系统而言，这些制约因素足以破坏其启动过程和运行稳定性。  
3. **数据持久化配置不当：** Docker卷（Volume）的使用方式，特别是当项目文件和Docker数据卷的物理存储位于跨WSL2边界的Windows文件系统时，会极大地加剧I/O性能瓶颈，导致数据库操作超时，进而引发服务启动失败。

**解决方案概述：** 解决此问题的策略必须是整体性的，旨在构建一个稳定、高性能的本地开发环境。具体措施包括：通过引入明确的、基于服务真实就绪状态的健康检查和依赖条件，对docker-compose.yml文件进行全面加固；通过创建并配置.wslconfig文件，对WSL2的资源分配进行显式优化；以及遵循最佳实践，对Docker数据卷的管理和项目文件的存放位置进行规范，从而根除性能瓶颈。

---

## **II. Temporal服务“不健康”状态分析**

本章节将系统性地剖析temporal容器无法达到“健康”（healthy）状态的深层原因，从表面症状（健康检查失败）入手，层层递进，揭示其背后的技术根源。

### **2.1. 解构健康检查机制**

Docker的healthcheck机制是在docker-compose.yml或Dockerfile中定义的一条命令，Docker引擎会周期性地在容器内部执行该命令，以判断容器内应用进程的真实健康状况 1。该命令的退出状态码（Exit Code）是判断依据：退出码

0表示服务健康，而1则表示不健康 1。容器的健康状态会经历

starting、healthy和unhealthy的转变。

在Temporal的早期版本中，其官方Docker镜像并未内置HEALTHCHECK指令。这导致了一个普遍问题：容器虽然处于“运行中”（running）状态，但容器内的Temporal服务进程可能仍在初始化、尚未准备好接收请求，或者已经因故崩溃。这种状态不一致常常导致依赖Temporal服务的下游应用（如Worker或UI）连接失败 4。

为解决此问题，Temporal社区探索并形成了一套行之有效的健康检查方案，即利用Temporal的命令行工具tctl来探测服务的真实状态 4。实践证明，最高效且可靠的健康检查命令是

tctl \--address temporal:7233 workflow list。这条命令优于另一常用命令tctl cluster health，其关键优势在于workflow list操作不仅检查gRPC前端的连通性，还隐式地验证了默认命名空间（default namespace）是否已成功创建并可用。由于命名空间的创建是temporal服务（特别是使用auto-setup镜像时）初始化流程的最后几个关键步骤之一，成功执行workflow list标志着服务已达到完全可操作的状态，从而有效规避了集群看似“健康”但尚未完全就绪的竞争条件 4。

### **2.2. 根本原因分析：一个多层次的故障场景**

#### **竞争条件：服务启动顺序的谬误**

Temporal服务的正常启动强依赖于其底层数据存储服务的完全就绪。根据官方的docker-compose.yml文件，temporal服务依赖于postgresql（持久化存储）和elasticsearch（可见性存储）5。然而，

docker-compose中的depends\_on指令在默认情况下，仅仅保证了容器的启动顺序，即postgresql和elasticsearch容器会在temporal容器之前启动。它并不能保证被依赖服务的内部应用已经完成了初始化并准备好接受连接。

这个问题的关键在于Elasticsearch的启动特性。Elasticsearch进程启动后，会很快开始监听HTTP端口并响应网络请求，但此时其内部的集群分片可能仍在初始化，集群状态可能为“红色”（red）或“黄色”（yellow），尚未达到可稳定写入数据的“绿色”（green）或（在单节点开发环境中可接受的）“黄色”状态 6。一个简单的端口探测式健康检查会过早地判断Elasticsearch为“健康”，而实际上它还无法处理模式创建（schema creation）等复杂请求。

当temporal容器（特别是使用temporalio/auto-setup镜像）启动时，其入口脚本会立即尝试连接postgresql和elasticsearch，以自动创建和初始化所需的数据库模式（schema）7。如果此时Elasticsearch尚未完全就绪，Temporal的模式设置脚本将因连接失败或集群状态不佳而超时或报错。这个初始化失败会直接导致

temporal服务进程异常退出或挂起，最终使得后续的tctl健康检查命令执行失败，容器状态被标记为unhealthy。

因此，**第一个根本原因**是由于对依赖服务（尤其是Elasticsearch）缺乏严格的、基于应用层就绪状态的健康检查，从而引发了致命的启动竞争条件，导致了连锁故障。

#### **资源枯竭：WSL2的内存陷阱**

在Windows上运行Docker Desktop时，其后端引擎实际上是在一个轻量级的WSL2虚拟机中运行的。这个虚拟机有其独立的资源配额，包括CPU和内存 8。根据微软的默认设置，WSL2虚拟机被允许消耗高达主机总物理内存50%的资源 8。例如，在一台拥有16GB内存的开发机上，名为

vmmem的进程（代表WSL2虚拟机）可能会迅速占用8GB内存，并且在负载降低后往往不会及时有效地释放这些内存。

Temporal的默认Docker Compose部署栈中包含了Elasticsearch服务 11，而Elasticsearch是一个众所周知的内存消耗大户。当Temporal服务栈启动时，多个服务（特别是Elasticsearch）的内存需求会急剧上升。如果WSL2虚拟机的内存使用量达到了其上限，WSL2内部的Linux内核会触发其内存不足杀手（Out-Of-Memory Killer, OOM Killer）机制。OOM Killer会选择性地终止内存占用最高的进程以回收内存，保护系统免于崩溃。在这种场景下，Elasticsearch容器内的Java进程无疑是首要目标。

一旦Elasticsearch进程被OOM Killer静默终止，temporal服务在启动时尝试连接它时就会失败。这种失败是突发且无明显日志的（从Docker层面看，容器可能只是简单地退出了），直接导致temporal服务初始化中断，并最终进入unhealthy状态。

因此，**第二个根本原因**是WSL2默认的无限制内存分配策略为内存密集型服务（如Elasticsearch）的稳定运行制造了一个充满风险的环境，资源争抢和静默崩溃是服务不健康的直接诱因。

#### **性能瓶颈：文件系统I/O的鸿沟**

WSL2架构内存在两种截然不同的文件系统：原生的Linux文件系统（如ext4）和通过9P协议挂载的Windows文件系统（NTFS），后者在WSL2内部的访问路径通常为/mnt/c/、/mnt/d/等 12。跨越这两个文件系统边界的I/O操作存在巨大的性能鸿沟。在Linux原生文件系统内的读写速度接近物理硬件的极限，而访问挂载的Windows驱动器上的文件则会慢上几个数量级 13。

如果用户将Temporal项目（包含docker-compose.yml）存放在Windows的某个目录下（例如C:\\Users\\YourUser\\Projects\\temporal-app），然后在WSL2终端中通过/mnt/c/Users/...路径来执行docker compose up，那么就会触发这个性能陷阱。具体体现在：

1. **绑定挂载（Bind Mounts）：** 像./dynamicconfig:/etc/temporal/config/dynamicconfig这样的配置，会直接将Windows文件系统上的目录映射到容器内，所有对该目录的读写都将承受巨大的性能损失。  
2. **命名卷（Named Volumes）：** 更为隐蔽的是，Docker Desktop在WSL2模式下创建的命名卷，其数据文件的物理存储位置与Docker守护进程的上下文环境有关。如果操作的上下文在Windows文件系统，那么为PostgreSQL和Elasticsearch创建的数据卷的性能也可能受到影响，从而严重拖慢数据库的事务提交、索引写入等关键I/O密集型操作。

对于Temporal服务而言，启动过程中的数据库模式初始化和运行时的健康检查都依赖于与PostgreSQL的快速、可靠通信。缓慢的磁盘I/O会导致数据库事务超时，进而使得Temporal的启动脚本或健康检查命令失败。

因此，**第三个根本原因**是将项目文件或Docker数据卷放置在Windows文件系统上，这是WSL2环境下使用Docker的一个严重的反模式（anti-pattern）。由此产生的I/O延迟是导致服务启动超时和健康检查失败的重要因素。

#### **网络问题：特定环境下的连接不稳定性**

WSL2的默认网络模式为NAT（网络地址转换），它在WSL2虚拟机和Windows主机之间建立了一个虚拟网络。在大多数情况下，这种模式工作良好。然而，在一些复杂的网络环境中，特别是企业内部网络，当主机连接到VPN或配置了严格的防火墙规则时，WSL2的网络连接可能会变得不稳定 16。

在Docker Compose创建的自定义桥接网络中，服务间的通信依赖于Docker内建的DNS解析器来解析服务名（例如，temporal服务通过主机名postgresql来访问数据库）19。在WSL2网络受外部因素（如VPN客户端修改主机路由表）干扰时，Docker的内部DNS解析可能会失败或出现显著延迟 21。一个明确的失败信号是在

temporal容器的启动日志中看到类似nc: bad address 'postgresql'的错误，这直接表明DNS解析失败，服务无法建立数据库连接，启动过程被迫中止 20。

因此，**第四个（较为次要的）根本原因**是，在特定网络环境下，WSL2的网络复杂性可能干扰Docker容器间的DNS解析，导致服务连接失败。

---

## **III. Temporal UI 500 错误调查**

本章节将深入分析Temporal UI返回500 Internal Server Error的原因，并阐明其与后端服务健康状态的直接联系，同时也会探讨其他可能的次要因素。

### **3.1. UI与后端的通信架构**

Temporal UI (temporal-ui 容器) 是一个纯粹的前端应用，它通过gRPC协议与Temporal后端服务（temporal 容器）进行通信，以获取工作流执行历史、命名空间信息、任务队列状态等数据 23。

这种通信链路的配置是通过temporal-ui服务定义中的TEMPORAL\_ADDRESS环境变量来完成的。该变量的值必须准确指向temporal服务的gRPC监听地址和端口，在Docker Compose环境中，这通常是服务名和端口号，例如temporal:7233 5。用户界面本身则默认在

8080端口上监听HTTP请求，供浏览器访问 5。用户在问题中提到的

8050端口很可能是一个笔误或自定义配置，但标准的默认端口是8080。

### **3.2. 根本原因分析：定位故障点**

#### **主要原因：后端服务不可用**

500 Internal Server Error是一个通用的HTTP状态码，它表明服务器在尝试处理请求时遇到了一个意外情况，导致无法完成请求 24。在Temporal UI的场景中，这个“服务器”就是

temporal-ui容器自身。

分析的逻辑链条非常清晰：

1. Temporal UI的核心功能是作为Temporal后端服务的数据可视化前端。  
2. 当用户在浏览器中访问UI并进行操作时，UI服务会向TEMPORAL\_ADDRESS所配置的后端gRPC端点发起请求。  
3. 正如第二章节所详尽分析的，temporal服务正处于unhealthy状态。这意味着其在temporal:7233上的gRPC服务要么没有启动，要么正在主动拒绝连接，要么在处理请求时内部崩溃。  
4. 当UI服务发出的gRPC请求因为网络错误、连接被拒或超时而失败时，其内部的服务器逻辑未能妥善处理这种与关键后端的通信中断，而是将这个底层错误向上冒泡，最终以一个通用的500错误响应返回给了用户的浏览器。

结论是显而易见的：**UI的500错误是后端服务不可用这一根本问题的直接、必然的表象。** 它不是UI本身的问题，而是其所依赖的“数据大脑”已经失能。因此，一旦解决了后端服务的健康问题，UI的500错误几乎必然会随之消失。这一点在一个相似的用户报告中得到了印证，该报告指出UI的500错误正是由于docker-compose.yml中后端依赖配置错误所导致的 26。

#### **次要原因：外部API调用的阻塞**

除了后端依赖问题，还存在一个潜在的次要原因。Temporal UI为了提供更好的用户体验，默认会启动一项功能：检查是否有新版本发布。这个检查是通过向GitHub的API端点 api.github.com 发起一个出站网络请求来实现的 27。

在某些网络受限的环境中——这在需要使用VPN或配置了严格防火墙的企业开发环境中非常常见，而这些环境也正是WSL2网络问题频发的地方——这个对外的API调用可能会被阻止。被阻止的请求可能会导致连接超时或收到403 Forbidden之类的HTTP错误。如果UI服务的早期版本对这类非关键性功能的网络故障处理不够优雅，也可能导致整个请求处理链路中断，从而向用户返回500错误。

尽管这个问题在较新版本的UI中已得到修复，但作为一个防御性措施，可以主动禁用此功能。通过在temporal-ui服务的环境变量中设置TEMPORAL\_NOTIFY\_ON\_NEW\_VERSION=false，可以完全避免此类潜在问题 27。

因此，尽管后端不健康是500错误的最主要原因，但在特定的受限网络环境中，版本检查功能的网络阻塞也可能是一个独立的或加剧问题的因素。

---

## **IV. 整体性加固解决方案**

本章节旨在提供一套完整、详尽且可直接应用的解决方案，以系统性地解决前述所有已识别的根本原因，从而构建一个稳定可靠的Temporal本地开发环境。

### **4.1. 打造一个高弹性的 docker-compose.yml**

以下是一个经过全面加固和详细注解的docker-compose.yml文件。它集成了所有最佳实践，旨在消除竞争条件、确保正确的启动顺序并实现可靠的数据持久化。

YAML

version: '3.8'

\# 定义顶级卷，确保数据持久化  
\# 使用命名卷是最佳实践，它们由Docker管理，独立于容器生命周期  
\# \[28, 29\]  
volumes:  
  postgresql\_data:  
    driver: local  
  elasticsearch\_data:  
    driver: local

networks:  
  temporal-network:  
    driver: bridge  
    name: temporal-network

services:  
  postgresql:  
    image: postgres:13  
    container\_name: temporal-postgresql  
    ports:  
      \- "5432:5432"  
    environment:  
      \- POSTGRES\_USER=temporal  
      \- POSTGRES\_PASSWORD=temporal  
      \- POSTGRES\_DB=temporal  
    volumes:  
      \# 将命名卷挂载到PostgreSQL的数据目录  
      \- postgresql\_data:/var/lib/postgresql/data  
    networks:  
      \- temporal-network  
    healthcheck:  
      \# 使用pg\_isready工具进行健康检查，确保数据库不仅启动，而且准备好接受连接  
      test:  
      interval: 5s  
      timeout: 5s  
      retries: 10

  elasticsearch:  
    image: elasticsearch:7.16.2  
    container\_name: temporal-elasticsearch  
    ports:  
      \- "9200:9200"  
      \- "9300:9300"  
    environment:  
      \- discovery.type=single-node  
      \# 为Java虚拟机分配适量的内存，防止资源滥用  
      \- "ES\_JAVA\_OPTS=-Xms1g \-Xmx1g"  
    volumes:  
      \# 将命名卷挂载到Elasticsearch的数据目录  
      \- elasticsearch\_data:/usr/share/elasticsearch/data  
    networks:  
      \- temporal-network  
    healthcheck:  
      \# 这是关键的加固措施：不仅仅是检查端口，而是查询集群健康API  
      \# 等待集群状态至少为 'yellow'，这在单节点环境中表示主分片已可用 \[6\]  
      test:  
      interval: 10s  
      timeout: 10s  
      retries: 12

  temporal:  
    image: temporalio/auto-setup:1.21.3  
    container\_name: temporal  
    ports:  
      \- "7233:7233"  
    environment:  
      \- DB=postgresql  
      \- DB\_PORT=5432  
      \- POSTGRES\_USER=temporal  
      \- POSTGRES\_PWD=temporal  
      \# 使用服务名进行DNS解析  
      \- POSTGRES\_SEEDS=postgresql  
      \- DYNAMIC\_CONFIG\_FILE\_PATH=config/dynamicconfig/development.yaml  
      \- ENABLE\_ES=true  
      \- ES\_SEEDS=elasticsearch  
      \- ES\_VERSION=v7  
    networks:  
      \- temporal-network  
    depends\_on:  
      \# 强制依赖：temporal必须在postgresql和elasticsearch都达到 'service\_healthy' 状态后才能启动  
      \# 这彻底解决了启动时的竞争条件问题 \[2, 4\]  
      postgresql:  
        condition: service\_healthy  
      elasticsearch:  
        condition: service\_healthy  
    healthcheck:  
      \# 使用最可靠的tctl命令作为健康检查，确保命名空间也已就绪 \[4\]  
      test:  
      interval: 10s  
      timeout: 5s  
      retries: 10  
      start\_period: 10s

  temporal-ui:  
    image: temporalio/ui:2.13.2  
    container\_name: temporal-ui  
    ports:  
      \- "8080:8080"  
    environment:  
      \# UI连接到后端的gRPC地址  
      \- TEMPORAL\_ADDRESS=temporal:7233  
      \# 防御性措施：禁用外部版本检查，避免在受限网络中出现问题 \[27\]  
      \- TEMPORAL\_NOTIFY\_ON\_NEW\_VERSION=false  
    networks:  
      \- temporal-network  
    depends\_on:  
      \# UI必须在temporal服务健康后才能启动  
      temporal:  
        condition: service\_healthy

### **4.2. 为Docker优化WSL2环境**

#### **.wslconfig 配置文件**

为了从根本上解决WSL2的资源滥用和性能问题，必须创建一个.wslconfig文件来显式地控制其虚拟机的资源分配。

**创建步骤：**

1. 打开Windows文件资源管理器。  
2. 在地址栏输入 %UserProfile% 并按回车。这将直接跳转到您的用户主目录（通常是 C:\\Users\\\<YourUsername\>）。  
3. 在此目录下，创建一个名为 .wslconfig 的新文件（注意文件名前面有一个点）。  
4. 使用文本编辑器打开该文件，并填入以下推荐配置。

**推荐配置：**

| 键 (Key) | 示例值 | 描述 |
| :---- | :---- | :---- |
| memory | 8GB | 限制WSL2虚拟机最多使用8GB物理内存。此值应根据您的主机总内存进行调整（建议不超过主机内存的一半），以避免资源争抢 10。 |
| processors | 4 | 限制WSL2虚拟机最多使用4个CPU逻辑核心。此值应根据您的主机CPU核心数进行调整（建议不超过主机核心数的一半）10。 |
| swap | 0 | 禁用交换文件。这可以防止在内存压力大时发生缓慢的磁盘交换，强制应用在内存不足时快速失败，而不是变得极其缓慢 10。 |
| networkingMode | NAT | 默认的网络模式。对于遇到由VPN或复杂防火墙引起的DNS解析问题的用户，可以尝试设置为实验性的mirrored模式，这有望改善网络兼容性 16。 |

**示例 .wslconfig 文件内容：**

Ini, TOML

\[wsl2\]  
memory\=8GB  
processors\=4  
swap\=0

应用配置：  
在创建或修改.wslconfig文件后，必须完全关闭WSL2虚拟机才能使配置生效。打开Windows PowerShell或命令提示符，执行以下命令：  
wsl \--shutdown  
之后，重新启动Docker Desktop或任何WSL2发行版，新的资源限制就会被应用。

#### **项目与数据卷的存放位置**

为了根除文件系统I/O瓶颈，必须遵循以下黄金法则：

**将您的整个Temporal项目，包括docker-compose.yml文件和所有相关代码，克隆或存放在WSL2的原生Linux文件系统内。**

一个推荐的路径是 /home/\<your-wsl-username\>/projects/temporal-project。**绝对不要**从挂载的Windows路径（如 /mnt/c/Users/...）中运行docker compose up命令 12。

当您在WSL2的Linux文件系统内操作时，Docker Desktop会自动将其创建的命名卷（如postgresql\_data）的数据存储在WSL2自己的虚拟硬盘（ext4.vhdx）内。这确保了所有数据库和搜索引擎的I/O操作都在高性能的Linux文件系统内部完成，从而获得接近本机的性能。

---

## **V. 高级诊断与长期维护**

本章节旨在提供一套实用的诊断工具和方法，帮助您在部署加固方案后，能够持续监控系统状态，并独立排查未来可能出现的任何问题。

### **5.1. 实用的日志分析指南**

日志是诊断分布式系统问题的首要窗口。

* **实时聚合日志：** 要同时查看所有服务的实时日志输出，请在项目目录下运行 docker compose logs \-f。这对于观察服务启动顺序、依赖等待以及发现第一个抛出错误的组件非常有用。  
* **查看单个服务日志：** 如果怀疑特定服务有问题，可以单独查看其日志，例如 docker compose logs \-f temporal 或 docker compose logs \-f elasticsearch。  
* **关键日志信息：** 在启动过程中，留意以下日志模式：  
  * temporal-postgresql 和 temporal-elasticsearch 的日志中会显示它们的健康检查命令成功执行。  
  * temporal 服务的日志会先显示等待依赖服务就绪，然后开始进行数据库模式设置，最后成功启动。  
  * temporal-ui 的日志会显示它成功连接到 temporal:7233。

### **5.2. 掌握核心诊断命令**

下表汇总了在排查此类问题时最关键的Docker和Temporal命令，掌握它们将使您能够快速定位问题。

**表1：核心诊断命令参考**

| 命令 | 示例 | 用途 |
| :---- | :---- | :---- |
| docker ps \-a | docker ps \-a | 查看所有容器的状态，包括已停止容器的退出码（Exit Code），有助于判断容器是正常停止还是因错误退出。 |
| docker inspect | docker inspect \--format='{{json.State.Health}}' temporal | **（极其重要）** 获取特定容器的详细健康检查日志。这会显示最近几次健康检查的命令输出和退出码，是诊断“unhealthy”状态的根本原因的最直接方法 1。 |
| docker stats | docker stats | 实时监控所有运行中容器的CPU、内存、网络I/O和磁盘I/O使用情况。这是诊断WSL2资源瓶颈、确认.wslconfig配置是否生效的关键工具 31。 |
| tctl cluster health | docker compose exec temporal tctl cluster health | 从容器内部执行，快速检查Temporal集群前端服务的基本健康状况和连通性 4。 |
| tctl namespace list | docker compose exec temporal tctl \--ns default namespace list | 验证默认命名空间是否已成功创建和注册，是比cluster health更全面的服务就绪状态检查。 |

---

## **VI. 结论**

本次深入分析揭示了在WSL2的Docker环境中部署Temporal时遇到的服务“不健康”和UI 500错误，其根源并非单一故障，而是一个由**配置、环境和性能**三个维度的问题交织而成的系统性故障。

**核心结论如下：**

* **竞争条件是直接导火索：** 默认的Docker Compose配置缺乏对服务间依赖就绪状态的严格保证，导致Temporal服务在数据库和搜索引擎尚未完全准备好时过早地进行初始化，从而引发启动失败。  
* **WSL2环境是问题的温床：** 未经优化的WSL2环境，以其默认的、贪婪的内存分配策略和跨文件系统的巨大I/O性能惩罚，为Temporal这样复杂的有状态应用提供了一个极其不稳定的运行基础。资源枯竭和I/O超时是导致服务进程崩溃和启动失败的深层原因。  
* **UI错误是后端故障的必然结果：** Temporal UI的500错误是其所依赖的后端服务处于“不健康”状态的直接反映，而非UI应用本身的缺陷。

本报告提出的**整体性解决方案**——通过一份加固的docker-compose.yml文件、一个精心调优的.wslconfig配置文件，以及遵循将项目置于Linux文件系统内的核心实践——共同构建了一个稳定、高性能且资源可控的开发平台。这套方案从根本上解决了上述所有问题，创建了一个可预测且可靠的环境。

最终，通过实施本报告中详述的加固措施和诊断方法，开发者可以充满信心地在WSL2上利用Docker部署和开发Temporal应用，将精力集中于业务逻辑的实现，而非基础设施的反复调试。

#### **引用的著作**

1. Discovery service healthcheck fails. Service unhealthy \- Docker Community Forums, 访问时间为 七月 24, 2025， [https://forums.docker.com/t/discovery-service-healthcheck-fails-service-unhealthy/148558](https://forums.docker.com/t/discovery-service-healthcheck-fails-service-unhealthy/148558)  
2. Docker compose container doesn't stop when healthcheck is supposed to fail, 访问时间为 七月 24, 2025， [https://forums.docker.com/t/docker-compose-container-doesnt-stop-when-healthcheck-is-supposed-to-fail/141352](https://forums.docker.com/t/docker-compose-container-doesnt-stop-when-healthcheck-is-supposed-to-fail/141352)  
3. How to View Docker-Compose Healthcheck Logs \- A Quick Guide | SigNoz, 访问时间为 七月 24, 2025， [https://signoz.io/guides/how-to-view-docker-compose-healthcheck-logs/](https://signoz.io/guides/how-to-view-docker-compose-healthcheck-logs/)  
4. Feature: implement docker healthcheck · Issue \#453 · temporalio/temporal \- GitHub, 访问时间为 七月 24, 2025， [https://github.com/temporalio/temporal/issues/453](https://github.com/temporalio/temporal/issues/453)  
5. raw.githubusercontent.com, 访问时间为 七月 24, 2025， [https://raw.githubusercontent.com/temporalio/docker-compose/main/docker-compose.yml](https://raw.githubusercontent.com/temporalio/docker-compose/main/docker-compose.yml)  
6. How to Implement Elasticsearch Health Check in Docker Compose | Baeldung on Ops, 访问时间为 七月 24, 2025， [https://www.baeldung.com/ops/elasticsearch-docker-compose](https://www.baeldung.com/ops/elasticsearch-docker-compose)  
7. temporalio/auto-setup \- Docker Image, 访问时间为 七月 24, 2025， [https://hub.docker.com/r/temporalio/auto-setup](https://hub.docker.com/r/temporalio/auto-setup)  
8. How to configure memory limits in WSL2 \- Willem's Fizzy Logic, 访问时间为 七月 24, 2025， [https://fizzylogic.nl/2023/01/05/how-to-configure-memory-limits-in-wsl2](https://fizzylogic.nl/2023/01/05/how-to-configure-memory-limits-in-wsl2)  
9. Get started with Docker remote containers on WSL 2 \- Learn Microsoft, 访问时间为 七月 24, 2025， [https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-containers](https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-containers)  
10. Advanced settings configuration in WSL | Microsoft Learn, 访问时间为 七月 24, 2025， [https://learn.microsoft.com/en-us/windows/wsl/wsl-config](https://learn.microsoft.com/en-us/windows/wsl/wsl-config)  
11. Temporal docker-compose files \- GitHub, 访问时间为 七月 24, 2025， [https://github.com/temporalio/docker-compose](https://github.com/temporalio/docker-compose)  
12. Working across Windows and Linux file systems \- Learn Microsoft, 访问时间为 七月 24, 2025， [https://learn.microsoft.com/en-us/windows/wsl/filesystems](https://learn.microsoft.com/en-us/windows/wsl/filesystems)  
13. Increase WSL2 and Docker Performance on Windows By 20x | by Suyash Singh | Medium, 访问时间为 七月 24, 2025， [https://medium.com/@suyashsingh.stem/increase-docker-performance-on-windows-by-20x-6d2318256b9a](https://medium.com/@suyashsingh.stem/increase-docker-performance-on-windows-by-20x-6d2318256b9a)  
14. Poor performance with docker desktop on windows in WSLv2 · Issue \#12401 \- GitHub, 访问时间为 七月 24, 2025， [https://github.com/docker/for-win/issues/12401](https://github.com/docker/for-win/issues/12401)  
15. Initial Impressions of WSL 2 | Hacker News, 访问时间为 七月 24, 2025， [https://news.ycombinator.com/item?id=23090143](https://news.ycombinator.com/item?id=23090143)  
16. Fixing some issues with WSL2 and Wireguard | GinkCode, 访问时间为 七月 24, 2025， [https://www.ginkcode.com/post/fixing-some-issues-with-wsl2-and-wireguard](https://www.ginkcode.com/post/fixing-some-issues-with-wsl2-and-wireguard)  
17. Troubleshooting Windows Subsystem for Linux | Microsoft Learn, 访问时间为 七月 24, 2025， [https://learn.microsoft.com/en-us/windows/wsl/troubleshooting](https://learn.microsoft.com/en-us/windows/wsl/troubleshooting)  
18. Why is there no network connectivity in Ubuntu using WSL 2 behind VPN? \- Super User, 访问时间为 七月 24, 2025， [https://superuser.com/questions/1582623/why-is-there-no-network-connectivity-in-ubuntu-using-wsl-2-behind-vpn](https://superuser.com/questions/1582623/why-is-there-no-network-connectivity-in-ubuntu-using-wsl-2-behind-vpn)  
19. Does this project run? Can I get paid support to set up self-hosted? \- Server Deployment, 访问时间为 七月 24, 2025， [https://community.temporal.io/t/does-this-project-run-can-i-get-paid-support-to-set-up-self-hosted/17240](https://community.temporal.io/t/does-this-project-run-can-i-get-paid-support-to-set-up-self-hosted/17240)  
20. docker-compose.yml does not start PostgreSQL · Issue \#2322 · temporalio/temporal \- GitHub, 访问时间为 七月 24, 2025， [https://github.com/temporalio/temporal/issues/2322](https://github.com/temporalio/temporal/issues/2322)  
21. Has there been an official fix, or acknowledgement of the networking issues with WSL2 and VPNs? (ex. SSH Timeout Issue) \- Reddit, 访问时间为 七月 24, 2025， [https://www.reddit.com/r/bashonubuntuonwindows/comments/1jdrk7w/has\_there\_been\_an\_official\_fix\_or\_acknowledgement/](https://www.reddit.com/r/bashonubuntuonwindows/comments/1jdrk7w/has_there_been_an_official_fix_or_acknowledgement/)  
22. No internet connection Ubuntu-WSL while VPN \- Super User, 访问时间为 七月 24, 2025， [https://superuser.com/questions/1630487/no-internet-connection-ubuntu-wsl-while-vpn](https://superuser.com/questions/1630487/no-internet-connection-ubuntu-wsl-while-vpn)  
23. temporalio/ui: Temporal UI \- GitHub, 访问时间为 七月 24, 2025， [https://github.com/temporalio/ui](https://github.com/temporalio/ui)  
24. When do we get 504, 429, 500 errors in Temporal UI? \- Server Deployment, 访问时间为 七月 24, 2025， [https://community.temporal.io/t/when-do-we-get-504-429-500-errors-in-temporal-ui/11257](https://community.temporal.io/t/when-do-we-get-504-429-500-errors-in-temporal-ui/11257)  
25. Help needed with "500: Internal Error" : r/OpenWebUI \- Reddit, 访问时间为 七月 24, 2025， [https://www.reddit.com/r/OpenWebUI/comments/1hfilp5/help\_needed\_with\_500\_internal\_error/](https://www.reddit.com/r/OpenWebUI/comments/1hfilp5/help_needed_with_500_internal_error/)  
26. \[Bug\] devcontainer not running appropriately · Issue \#92 · temporalio/samples-typescript, 访问时间为 七月 24, 2025， [https://github.com/temporalio/samples-typescript/issues/92](https://github.com/temporalio/samples-typescript/issues/92)  
27. Disable Releases checking on UI \- Community Support \- Temporal, 访问时间为 七月 24, 2025， [https://community.temporal.io/t/disable-releases-checking-on-ui/6488](https://community.temporal.io/t/disable-releases-checking-on-ui/6488)  
28. Use volumes \- Docker Docs, 访问时间为 七月 24, 2025， [https://docs.docker.com/storage/volumes/](https://docs.docker.com/storage/volumes/)  
29. How to persist data in a dockerized postgres database using volumes? \- DEV Community, 访问时间为 七月 24, 2025， [https://dev.to/iamrj846/how-to-persist-data-in-a-dockerized-postgres-database-using-volumes-15f0](https://dev.to/iamrj846/how-to-persist-data-in-a-dockerized-postgres-database-using-volumes-15f0)  
30. container always end up 'unhealthy'? : r/docker \- Reddit, 访问时间为 七月 24, 2025， [https://www.reddit.com/r/docker/comments/1ep0tol/container\_always\_end\_up\_unhealthy/](https://www.reddit.com/r/docker/comments/1ep0tol/container_always_end_up_unhealthy/)  
31. Advanced Container Resource Monitoring with docker stats \- Last9, 访问时间为 七月 24, 2025， [https://last9.io/blog/container-resource-monitoring-with-docker-stats/](https://last9.io/blog/container-resource-monitoring-with-docker-stats/)  
32. How to Monitor Container Memory and CPU Usage in Docker Desktop, 访问时间为 七月 24, 2025， [https://www.docker.com/blog/how-to-monitor-container-memory-and-cpu-usage-in-docker-desktop/](https://www.docker.com/blog/how-to-monitor-container-memory-and-cpu-usage-in-docker-desktop/)  
33. tctl v1.17 cluster command reference | Temporal Platform Documentation, 访问时间为 七月 24, 2025， [https://docs.temporal.io/tctl-v1/cluster](https://docs.temporal.io/tctl-v1/cluster)