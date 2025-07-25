#!/usr/bin/env node

// 前端JavaScript功能单元测试 (Node.js环境)
console.log('🏰 Cube Castle - 前端JavaScript单元测试');
console.log('========================================');

// 测试结果记录
let testResults = [];

// 测试辅助函数
function assert(condition, message) {
    if (condition) {
        testResults.push({ pass: true, message: `✅ ${message}` });
        console.log(`✅ ${message}`);
    } else {
        testResults.push({ pass: false, message: `❌ ${message}` });
        console.log(`❌ ${message}`);
    }
}

function assertEqual(actual, expected, message) {
    assert(actual === expected, `${message} - 期望: ${expected}, 实际: ${actual}`);
}

function assertNotNull(value, message) {
    assert(value !== null && value !== undefined, `${message} - 值不应为空`);
}

function assertType(value, expectedType, message) {
    assert(typeof value === expectedType, `${message} - 期望类型: ${expectedType}, 实际类型: ${typeof value}`);
}

// 测试1: 测试JavaScript基础功能
function testJavaScriptBasics() {
    console.log('\n开始测试JavaScript基础功能...');
    
    // 测试数据类型
    assertEqual(typeof 'test', 'string', '测试字符串类型');
    assertEqual(typeof 123, 'number', '测试数字类型');
    assertEqual(typeof true, 'boolean', '测试布尔类型');
    assertEqual(typeof {}, 'object', '测试对象类型');
    assertEqual(typeof [], 'object', '测试数组类型');
    assertEqual(typeof function(){}, 'function', '测试函数类型');
    
    // 测试数组操作
    const arr = [1, 2, 3];
    assertEqual(arr.length, 3, '数组长度');
    arr.push(4);
    assertEqual(arr.length, 4, '数组push操作');
    assertEqual(arr[3], 4, '数组元素访问');
}

// 测试2: 测试JSON处理功能
function testJSONHandling() {
    console.log('\n开始测试JSON处理...');
    
    const testData = {
        employee_number: 'EMP12345',
        first_name: '张三',
        last_name: '李',
        email: 'zhangsan@example.com',
        hire_date: '2024-01-01'
    };
    
    // 测试JSON序列化
    let jsonString;
    try {
        jsonString = JSON.stringify(testData);
        assert(true, 'JSON序列化成功');
        assertType(jsonString, 'string', 'JSON序列化结果类型');
    } catch (error) {
        assert(false, `JSON序列化失败: ${error.message}`);
    }
    
    // 测试JSON反序列化
    try {
        const parsed = JSON.parse(jsonString);
        assert(true, 'JSON反序列化成功');
        assertEqual(parsed.employee_number, testData.employee_number, 'JSON反序列化数据正确性');
        assertEqual(parsed.email, testData.email, 'JSON反序列化邮箱字段');
        assertEqual(parsed.first_name, testData.first_name, 'JSON反序列化中文字段');
    } catch (error) {
        assert(false, `JSON反序列化失败: ${error.message}`);
    }
}

// 测试3: 测试表单验证功能
function testFormValidation() {
    console.log('\n开始测试表单验证...');
    
    // 模拟表单验证函数
    function validateEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }
    
    function validateEmployeeName(name) {
        return name && name.trim().length > 0;
    }
    
    function validateEmployeeNumber(empNumber) {
        return empNumber && empNumber.startsWith('EMP') && empNumber.length > 3;
    }
    
    // 测试邮箱验证
    assert(validateEmail('test@example.com'), '有效邮箱验证通过');
    assert(validateEmail('user.name@company.com.cn'), '复杂有效邮箱验证通过');
    assert(!validateEmail('invalid-email'), '无效邮箱验证失败');
    assert(!validateEmail('test@'), '不完整邮箱验证失败');
    assert(!validateEmail(''), '空邮箱验证失败');
    
    // 测试姓名验证
    assert(validateEmployeeName('张三'), '有效姓名验证通过');
    assert(validateEmployeeName('John Doe'), '英文姓名验证通过');
    assert(!validateEmployeeName(''), '空姓名验证失败');
    assert(!validateEmployeeName('   '), '空白姓名验证失败');
    assert(!validateEmployeeName(null), 'null姓名验证失败');
    
    // 测试员工编号验证
    assert(validateEmployeeNumber('EMP12345'), '有效员工编号验证通过');
    assert(validateEmployeeNumber('EMP' + Date.now()), '动态员工编号验证通过');
    assert(!validateEmployeeNumber('ABC123'), '无效前缀员工编号验证失败');
    assert(!validateEmployeeNumber('EMP'), '太短员工编号验证失败');
    assert(!validateEmployeeNumber(''), '空员工编号验证失败');
}

// 测试4: 测试异步操作和Promise
async function testAsyncOperations() {
    console.log('\n开始测试异步操作...');
    
    // 测试Promise基础功能
    const promiseTest = await Promise.resolve('test');
    assertEqual(promiseTest, 'test', 'Promise resolve测试');
    
    // 测试async/await
    async function asyncFunction() {
        return '异步结果';
    }
    
    const asyncResult = await asyncFunction();
    assertEqual(asyncResult, '异步结果', 'async/await测试');
    
    // 测试setTimeout模拟
    await new Promise((resolve) => {
        setTimeout(() => {
            assert(true, '异步定时器操作成功');
            resolve();
        }, 50);
    });
}

// 测试5: 测试API请求数据格式
function testAPIRequestFormat() {
    console.log('\n开始测试API请求格式...');
    
    // 测试创建员工请求数据格式
    function createEmployeeRequestData(name, email) {
        const nameParts = name.split(' ');
        return {
            employee_number: 'EMP' + Date.now(),
            first_name: nameParts[0] || name,
            last_name: nameParts[1] || '',
            email: email,
            hire_date: new Date().toISOString().split('T')[0]
        };
    }
    
    const employeeData = createEmployeeRequestData('张三 李', 'zhangsan@example.com');
    
    // 验证数据格式
    assertType(employeeData.employee_number, 'string', '员工编号类型');
    assert(employeeData.employee_number.startsWith('EMP'), '员工编号前缀正确');
    assertType(employeeData.first_name, 'string', '名字类型');
    assertEqual(employeeData.first_name, '张三', '名字解析正确');
    assertEqual(employeeData.last_name, '李', '姓氏解析正确');
    assertType(employeeData.email, 'string', '邮箱类型');
    assert(employeeData.hire_date.match(/^\d{4}-\d{2}-\d{2}$/), '日期格式正确');
    
    // 测试单个名字的处理
    const singleNameData = createEmployeeRequestData('王五', 'wangwu@example.com');
    assertEqual(singleNameData.first_name, '王五', '单个名字处理正确');
    assertEqual(singleNameData.last_name, '', '单个名字时姓氏为空');
}

// 测试6: 测试错误处理
function testErrorHandling() {
    console.log('\n开始测试错误处理...');
    
    // 测试try-catch
    try {
        JSON.parse('invalid json');
        assert(false, 'JSON解析应该抛出错误');
    } catch (error) {
        assert(true, 'JSON解析错误被正确捕获');
        assertType(error.message, 'string', '错误消息类型');
    }
    
    // 测试空值处理
    function safeAccess(obj, key) {
        try {
            return obj && obj[key] ? obj[key] : null;
        } catch (error) {
            return null;
        }
    }
    
    assertEqual(safeAccess({name: 'test'}, 'name'), 'test', '正常属性访问');
    assertEqual(safeAccess(null, 'name'), null, '空对象安全访问');
    assertEqual(safeAccess({}, 'name'), null, '不存在属性安全访问');
}

// 运行所有测试
async function runAllTests() {
    console.log('开始运行前端JavaScript单元测试...\n');
    
    try {
        testJavaScriptBasics();
        testJSONHandling();
        testFormValidation();
        await testAsyncOperations();
        testAPIRequestFormat();
        testErrorHandling();
        
        // 统计测试结果
        let passCount = 0;
        let failCount = 0;
        
        testResults.forEach(result => {
            if (result.pass) {
                passCount++;
            } else {
                failCount++;
            }
        });
        
        // 显示测试总结
        const total = passCount + failCount;
        console.log('\n========================================');
        console.log('前端JavaScript单元测试完成！');
        console.log(`总计: ${total} 项测试`);
        console.log(`✅ 通过: ${passCount} 项`);
        console.log(`❌ 失败: ${failCount} 项`);
        console.log(`成功率: ${(passCount / total * 100).toFixed(1)}%`);
        console.log('========================================');
        
        return { total, passCount, failCount };
        
    } catch (error) {
        console.error(`测试执行出错: ${error.message}`);
        return { total: 0, passCount: 0, failCount: 1 };
    }
}

// 如果是直接运行此脚本
if (require.main === module) {
    runAllTests();
}

module.exports = { runAllTests };