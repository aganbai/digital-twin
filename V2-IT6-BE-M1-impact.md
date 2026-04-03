## 字段影响清单
1. models.go: Memory 结构体新增 MemoryLayer 字段。
2. database.go: DDL (alterMemoriesAddMemoryLayer), Index (createMemoriesLayerIndex, createMemoriesTypeLayerIndex), 及 autoMigrate 自动迁移。
3. repository_memory.go: 
   - Create / CreateWithPersonas 插入 memory_layer
   - GetByStudentAndTeacher / GetByPersonas 查询增加 memory_layer
   - 新增 UpdateMemoryLayer, GetByID 等方法
   - scanMemories 等辅助方法解析 memory_layer
4. 其他使用到 memory 的地方。
代码中大部分变更已存在，需要补全缺漏部分。
