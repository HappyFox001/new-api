# New API - TypeScript API 使用指南

本文档提供了使用 TypeScript 管理 API Key 完整生命周期的详细指南。

## 目录
- [环境配置](#环境配置)
- [API 接口定义](#api-接口定义)
- [1. 创建 API Key](#1-创建-api-key)
- [2. 查询 API Key 信息](#2-查询-api-key-信息)
- [3. 更新 API Key 额度](#3-更新-api-key-额度)
- [4. 增加 API Key 额度](#4-增加-api-key-额度)
- [5. 完整使用示例](#5-完整使用示例)
- [错误处理](#错误处理)

## 环境配置

首先安装必要的依赖：

```bash
npm install axios
npm install -D @types/node
```

创建基础配置文件：

```typescript
// config.ts
export const API_CONFIG = {
  baseURL: 'http://localhost:8000/api/auto-token',
  timeout: 80000,
};

// 如果你有统一的用户账户信息
export const UNIFIED_USER = {
  username: 'your_unified_username',
  password: 'your_unified_password',
};
```

## API 接口定义

```typescript
// types.ts

// Request Types
export interface CreateTokenRequest {
  username: string;
  password: string;
  token_name: string;
  remain_quota?: number;
  expired_time?: number;
  group?: string;
}

export interface UpdateTokenQuotaRequest {
  token_id: number;
  remain_quota: number;
}

export interface UpdateTokenQuotaByKeyRequest {
  api_key: string;
  remain_quota: number;
}

export interface AddTokenQuotaRequest {
  api_key: string;
  add_quota: number;
}

export interface GetTokenInfoRequest {
  api_key: string;
}

// Response Types
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
}

export interface CreateTokenData {
  token_id: number;
  key: string;
  user_id: number;
}

export interface TokenInfo {
  token_id: number;
  name: string;
  remain_quota: number;
  used_quota: number;
  created_time: number;
  expired_time: number;
  group: string;
  status: number;
}

export type CreateTokenResponse = ApiResponse<CreateTokenData>;
export type TokenInfoResponse = ApiResponse<TokenInfo>;
export type UpdateQuotaResponse = ApiResponse;
```

## API 客户端类

```typescript
// api-client.ts
import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
  CreateTokenRequest,
  CreateTokenResponse,
  UpdateTokenQuotaRequest,
  UpdateTokenQuotaByKeyRequest,
  AddTokenQuotaRequest,
  GetTokenInfoRequest,
  TokenInfoResponse,
  UpdateQuotaResponse,
  ApiResponse,
} from './types';
import { API_CONFIG } from './config';

export class NewApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string = API_CONFIG.baseURL) {
    this.client = axios.create({
      baseURL,
      timeout: API_CONFIG.timeout,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 添加响应拦截器用于错误处理
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('API Error:', error.response?.data || error.message);
        return Promise.reject(error);
      }
    );
  }

  /**
   * 创建新的 API Key
   */
  async createToken(request: CreateTokenRequest): Promise<CreateTokenResponse> {
    const response: AxiosResponse<CreateTokenResponse> = await this.client.post('/create', request);
    return response.data;
  }

  /**
   * 通过 Token ID 更新额度
   */
  async updateTokenQuota(request: UpdateTokenQuotaRequest): Promise<UpdateQuotaResponse> {
    const response: AxiosResponse<UpdateQuotaResponse> = await this.client.put('/quota', request);
    return response.data;
  }

  /**
   * 通过 API Key 更新额度
   */
  async updateTokenQuotaByKey(request: UpdateTokenQuotaByKeyRequest): Promise<UpdateQuotaResponse> {
    const response: AxiosResponse<UpdateQuotaResponse> = await this.client.put('/quota-by-key', request);
    return response.data;
  }

  /**
   * 为 API Key 增加额度
   */
  async addTokenQuota(request: AddTokenQuotaRequest): Promise<UpdateQuotaResponse> {
    const response: AxiosResponse<UpdateQuotaResponse> = await this.client.post('/add-quota', request);
    return response.data;
  }

  /**
   * 查询 API Key 信息
   */
  async getTokenInfo(request: GetTokenInfoRequest): Promise<TokenInfoResponse> {
    const response: AxiosResponse<TokenInfoResponse> = await this.client.post('/info', request);
    return response.data;
  }
}
```

## 1. 创建 API Key

```typescript
// create-token.ts
import { NewApiClient } from './api-client';
import { UNIFIED_USER } from './config';

async function createApiKeyForUser(userIdentifier: string, initialQuota: number = 100000): Promise<string | null> {
  const client = new NewApiClient();
  
  try {
    const response = await client.createToken({
      username: UNIFIED_USER.username,
      password: UNIFIED_USER.password,
      token_name: `Token for ${userIdentifier}`,
      remain_quota: initialQuota,
      expired_time: -1, // 永不过期
      group: 'default',
    });

    if (response.success) {
      console.log(`✅ API Key created successfully for ${userIdentifier}`);
      console.log(`🔑 API Key: ${response.data?.key}`);
      console.log(`🆔 Token ID: ${response.data?.token_id}`);
      console.log(`💰 Initial Quota: ${initialQuota}`);
      
      return response.data?.key || null;
    } else {
      console.error(`❌ Failed to create API Key: ${response.message}`);
      return null;
    }
  } catch (error) {
    console.error('❌ Error creating API Key:', error);
    return null;
  }
}

// 使用示例
async function example() {
  const apiKey = await createApiKeyForUser('user_12345', 500000);
  if (apiKey) {
    // 保存 API Key 到你的数据库
    console.log('API Key created and ready to use:', apiKey);
  }
}
```

## 2. 查询 API Key 信息

```typescript
// query-token.ts
import { NewApiClient } from './api-client';

async function getApiKeyInfo(apiKey: string) {
  const client = new NewApiClient();
  
  try {
    const response = await client.getTokenInfo({ api_key: apiKey });
    
    if (response.success && response.data) {
      const info = response.data;
      console.log('📊 API Key Information:');
      console.log(`🆔 Token ID: ${info.token_id}`);
      console.log(`📝 Name: ${info.name}`);
      console.log(`💰 Remaining Quota: ${info.remain_quota.toLocaleString()}`);
      console.log(`📈 Used Quota: ${info.used_quota.toLocaleString()}`);
      console.log(`📅 Created: ${new Date(info.created_time * 1000).toLocaleString()}`);
      console.log(`⏰ Expires: ${info.expired_time === -1 ? 'Never' : new Date(info.expired_time * 1000).toLocaleString()}`);
      console.log(`🏷️ Group: ${info.group}`);
      console.log(`🔄 Status: ${info.status === 1 ? 'Active' : 'Inactive'}`);
      
      return info;
    } else {
      console.error(`❌ Failed to get token info: ${response.message}`);
      return null;
    }
  } catch (error) {
    console.error('❌ Error getting token info:', error);
    return null;
  }
}

// 检查余额是否足够
async function checkQuotaSufficient(apiKey: string, requiredQuota: number): Promise<boolean> {
  const info = await getApiKeyInfo(apiKey);
  if (info) {
    const sufficient = info.remain_quota >= requiredQuota;
    console.log(`💳 Quota check: ${sufficient ? '✅ Sufficient' : '❌ Insufficient'}`);
    console.log(`💰 Required: ${requiredQuota.toLocaleString()}, Available: ${info.remain_quota.toLocaleString()}`);
    return sufficient;
  }
  return false;
}
```

## 3. 更新 API Key 额度

```typescript
// update-quota.ts
import { NewApiClient } from './api-client';

/**
 * 设置 API Key 的新额度（覆盖原有额度）
 */
async function setApiKeyQuota(apiKey: string, newQuota: number): Promise<boolean> {
  const client = new NewApiClient();
  
  try {
    const response = await client.updateTokenQuotaByKey({
      api_key: apiKey,
      remain_quota: newQuota,
    });
    
    if (response.success) {
      console.log(`✅ Quota updated successfully`);
      console.log(`💰 New quota: ${newQuota.toLocaleString()}`);
      return true;
    } else {
      console.error(`❌ Failed to update quota: ${response.message}`);
      return false;
    }
  } catch (error) {
    console.error('❌ Error updating quota:', error);
    return false;
  }
}

/**
 * 批量更新多个 API Key 的额度
 */
async function batchUpdateQuotas(updates: Array<{ apiKey: string; quota: number }>) {
  const client = new NewApiClient();
  const results = [];
  
  for (const update of updates) {
    try {
      const response = await client.updateTokenQuotaByKey({
        api_key: update.apiKey,
        remain_quota: update.quota,
      });
      
      results.push({
        apiKey: update.apiKey,
        success: response.success,
        message: response.message,
      });
      
      // 避免请求过于频繁，添加小延迟
      await new Promise(resolve => setTimeout(resolve, 100));
    } catch (error) {
      results.push({
        apiKey: update.apiKey,
        success: false,
        message: error instanceof Error ? error.message : 'Unknown error',
      });
    }
  }
  
  console.log('📊 Batch update results:');
  results.forEach(result => {
    console.log(`${result.success ? '✅' : '❌'} ${result.apiKey}: ${result.message}`);
  });
  
  return results;
}
```

## 4. 增加 API Key 额度

```typescript
// add-quota.ts
import { NewApiClient } from './api-client';

/**
 * 为 API Key 充值额度
 */
async function rechargeApiKey(apiKey: string, amount: number): Promise<boolean> {
  const client = new NewApiClient();
  
  try {
    // 先查询当前额度
    const currentInfo = await client.getTokenInfo({ api_key: apiKey });
    if (!currentInfo.success || !currentInfo.data) {
      console.error('❌ Cannot get current token info');
      return false;
    }
    
    const beforeQuota = currentInfo.data.remain_quota;
    
    // 增加额度
    const response = await client.addTokenQuota({
      api_key: apiKey,
      add_quota: amount,
    });
    
    if (response.success) {
      console.log(`✅ Quota recharged successfully`);
      console.log(`💰 Before: ${beforeQuota.toLocaleString()}`);
      console.log(`➕ Added: ${amount.toLocaleString()}`);
      console.log(`💰 After: ${(beforeQuota + amount).toLocaleString()}`);
      return true;
    } else {
      console.error(`❌ Failed to recharge quota: ${response.message}`);
      return false;
    }
  } catch (error) {
    console.error('❌ Error recharging quota:', error);
    return false;
  }
}

/**
 * 自动充值 - 当余额低于阈值时自动充值
 */
async function autoRecharge(apiKey: string, threshold: number, rechargeAmount: number): Promise<boolean> {
  const client = new NewApiClient();
  
  try {
    const info = await client.getTokenInfo({ api_key: apiKey });
    if (!info.success || !info.data) {
      console.error('❌ Cannot get token info for auto recharge');
      return false;
    }
    
    const currentQuota = info.data.remain_quota;
    console.log(`💰 Current quota: ${currentQuota.toLocaleString()}`);
    console.log(`⚠️ Threshold: ${threshold.toLocaleString()}`);
    
    if (currentQuota < threshold) {
      console.log(`🔄 Auto recharge triggered`);
      return await rechargeApiKey(apiKey, rechargeAmount);
    } else {
      console.log(`✅ Quota sufficient, no recharge needed`);
      return true;
    }
  } catch (error) {
    console.error('❌ Error in auto recharge:', error);
    return false;
  }
}
```

## 5. 完整使用示例

```typescript
// complete-example.ts
import { NewApiClient } from './api-client';
import { UNIFIED_USER } from './config';

/**
 * 完整的 API Key 生命周期管理示例
 */
class ApiKeyManager {
  private client: NewApiClient;
  private apiKeys: Map<string, string> = new Map(); // userId -> apiKey

  constructor() {
    this.client = new NewApiClient();
  }

  /**
   * 为新用户创建 API Key
   */
  async createUserApiKey(userId: string, initialQuota: number = 100000): Promise<string | null> {
    console.log(`🚀 Creating API Key for user: ${userId}`);
    
    try {
      const response = await this.client.createToken({
        username: UNIFIED_USER.username,
        password: UNIFIED_USER.password,
        token_name: `User-${userId}-Token`,
        remain_quota: initialQuota,
        expired_time: -1,
        group: 'user',
      });

      if (response.success && response.data) {
        const apiKey = response.data.key;
        this.apiKeys.set(userId, apiKey);
        
        console.log(`✅ API Key created for ${userId}: ${apiKey}`);
        return apiKey;
      } else {
        console.error(`❌ Failed to create API Key: ${response.message}`);
        return null;
      }
    } catch (error) {
      console.error('❌ Error creating API Key:', error);
      return null;
    }
  }

  /**
   * 用户充值
   */
  async userRecharge(userId: string, amount: number): Promise<boolean> {
    const apiKey = this.apiKeys.get(userId);
    if (!apiKey) {
      console.error(`❌ No API Key found for user: ${userId}`);
      return false;
    }

    console.log(`💳 Processing recharge for ${userId}: ${amount.toLocaleString()}`);
    
    try {
      const response = await this.client.addTokenQuota({
        api_key: apiKey,
        add_quota: amount,
      });

      if (response.success) {
        console.log(`✅ Recharge successful for ${userId}`);
        await this.getUserBalance(userId); // 显示新余额
        return true;
      } else {
        console.error(`❌ Recharge failed: ${response.message}`);
        return false;
      }
    } catch (error) {
      console.error('❌ Error processing recharge:', error);
      return false;
    }
  }

  /**
   * 查询用户余额
   */
  async getUserBalance(userId: string): Promise<number | null> {
    const apiKey = this.apiKeys.get(userId);
    if (!apiKey) {
      console.error(`❌ No API Key found for user: ${userId}`);
      return null;
    }

    try {
      const response = await this.client.getTokenInfo({ api_key: apiKey });
      
      if (response.success && response.data) {
        const balance = response.data.remain_quota;
        console.log(`💰 ${userId} balance: ${balance.toLocaleString()}`);
        return balance;
      } else {
        console.error(`❌ Failed to get balance: ${response.message}`);
        return null;
      }
    } catch (error) {
      console.error('❌ Error getting balance:', error);
      return null;
    }
  }

  /**
   * 监控低余额用户
   */
  async monitorLowBalance(threshold: number = 10000) {
    console.log(`🔍 Monitoring users with balance below ${threshold.toLocaleString()}`);
    
    const lowBalanceUsers = [];
    
    for (const [userId, apiKey] of this.apiKeys.entries()) {
      try {
        const response = await this.client.getTokenInfo({ api_key: apiKey });
        
        if (response.success && response.data) {
          const balance = response.data.remain_quota;
          if (balance < threshold) {
            lowBalanceUsers.push({ userId, balance, apiKey });
            console.log(`⚠️ Low balance: ${userId} has ${balance.toLocaleString()}`);
          }
        }
      } catch (error) {
        console.error(`❌ Error checking balance for ${userId}:`, error);
      }
    }
    
    return lowBalanceUsers;
  }

  /**
   * 批量充值
   */
  async batchRecharge(recharges: Array<{ userId: string; amount: number }>) {
    console.log(`📦 Processing batch recharge for ${recharges.length} users`);
    
    const results = [];
    
    for (const recharge of recharges) {
      const success = await this.userRecharge(recharge.userId, recharge.amount);
      results.push({
        userId: recharge.userId,
        amount: recharge.amount,
        success,
      });
      
      // 避免请求过于频繁
      await new Promise(resolve => setTimeout(resolve, 200));
    }
    
    console.log('📊 Batch recharge results:');
    results.forEach(result => {
      console.log(`${result.success ? '✅' : '❌'} ${result.userId}: ${result.amount.toLocaleString()}`);
    });
    
    return results;
  }
}

// 使用示例
async function demo() {
  const manager = new ApiKeyManager();
  
  // 1. 为用户创建 API Key
  await manager.createUserApiKey('user001', 200000);
  await manager.createUserApiKey('user002', 150000);
  
  // 2. 查询余额
  await manager.getUserBalance('user001');
  
  // 3. 用户充值
  await manager.userRecharge('user001', 50000);
  
  // 4. 监控低余额
  await manager.monitorLowBalance(100000);
  
  // 5. 批量充值
  await manager.batchRecharge([
    { userId: 'user001', amount: 80000 },
    { userId: 'user002', amount: 20000 },
  ]);
}

// 运行示例
// demo().catch(console.error);
```

## 错误处理

```typescript
// error-handler.ts

export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public originalError?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export function handleApiError(error: any): never {
  if (error.response) {
    // 服务器返回错误状态码
    const statusCode = error.response.status;
    const message = error.response.data?.message || 'API request failed';
    throw new ApiError(message, statusCode, error);
  } else if (error.request) {
    // 请求发送但没有收到响应
    throw new ApiError('No response from server', 0, error);
  } else {
    // 其他错误
    throw new ApiError(error.message || 'Unknown error', 0, error);
  }
}

// 在 API 客户端中使用
export class SafeNewApiClient extends NewApiClient {
  async createToken(request: CreateTokenRequest): Promise<CreateTokenResponse> {
    try {
      return await super.createToken(request);
    } catch (error) {
      handleApiError(error);
    }
  }

  // 其他方法类似...
}
```

## 环境变量配置

```typescript
// env.ts
export const ENV = {
  NEW_API_BASE_URL: process.env.NEW_API_BASE_URL || 'http://localhost:8000/api/auto-token',
  UNIFIED_USERNAME: process.env.UNIFIED_USERNAME || '',
  UNIFIED_PASSWORD: process.env.UNIFIED_PASSWORD || '',
  DEFAULT_QUOTA: parseInt(process.env.DEFAULT_QUOTA || '100000'),
  LOW_BALANCE_THRESHOLD: parseInt(process.env.LOW_BALANCE_THRESHOLD || '10000'),
};

// 验证必需的环境变量
export function validateEnv() {
  if (!ENV.UNIFIED_USERNAME || !ENV.UNIFIED_PASSWORD) {
    throw new Error('Missing required environment variables: UNIFIED_USERNAME, UNIFIED_PASSWORD');
  }
}
```

## 总结

这个 TypeScript 指南涵盖了 API Key 的完整生命周期：

1. **创建阶段**：使用统一账户为用户创建 API Key
2. **查询阶段**：获取 API Key 信息和余额
3. **更新阶段**：设置新的额度或增加额度
4. **监控阶段**：监控余额，自动充值
5. **批量操作**：批量管理多个 API Key

所有代码都包含了完整的类型定义、错误处理和实用的示例，可以直接在你的项目中使用。 