/**
 * Marketplace OpenAPI — Pre-request Script（HS256 JWT）
 *
 * 用法：
 * 1. 在 Environment 或 Collection 变量中配置：
 *    - jwt_account   与 conf 中 jwt.clients[].account 一致
 *    - jwt_secret    与对应 client 的 secret 一致
 *    - jwt_ttl_seconds（可选）默认 300，须 ≤ 服务端 jwt.expiration_seconds
 *    - jwt_payload_json（可选）业务字段 JSON 字符串（仅 /openapi/ecoin 等需 JWT 的接口）
 * 2. Body 选 raw / JSON，可留空或任意占位；脚本会覆盖为 {"jwt":"..."}
 * 3. Content-Type: application/json
 *
 * 依赖：Postman 内置 crypto-js（Sandbox）
 */
const CryptoJS = require('crypto-js');

function base64urlEncodeUtf8(str) {
    return CryptoJS.enc.Base64.stringify(CryptoJS.enc.Utf8.parse(str))
        .replace(/=/g, '')
        .replace(/\+/g, '-')
        .replace(/\//g, '_');
}

function hmacSha256Base64Url(data, secret) {
    return CryptoJS.HmacSHA256(data, secret)
        .toString(CryptoJS.enc.Base64)
        .replace(/=/g, '')
        .replace(/\+/g, '-')
        .replace(/\//g, '_');
}

function getVar(name) {
    const v = pm.environment.get(name);
    if (v !== undefined && v !== null && String(v) !== '') {
        return v;
    }
    return pm.collectionVariables.get(name);
}

const account = getVar('jwt_account');
const secret = getVar('jwt_secret');
if (!account || !secret) {
    throw new Error('请设置 jwt_account、jwt_secret（Environment 或 Collection 变量）');
}

const ttl = parseInt(getVar('jwt_ttl_seconds') || '300', 10);
const now = Math.floor(Date.now() / 1000);

/**
 * 环境变量里只能是字符串；若在本脚本内内联写成对象，也必须走此函数，禁止 JSON.parse(对象)。
 */
function parseExtraPayload(raw) {
    if (raw == null || raw === '') {
        return {};
    }
    if (typeof raw === 'object' && !Array.isArray(raw)) {
        return Object.assign({}, raw);
    }
    if (typeof raw === 'string') {
        try {
            return JSON.parse(raw);
        } catch (e) {
            throw new Error('jwt_payload_json 须为合法 JSON 字符串: ' + e.message);
        }
    }
    throw new Error('jwt_payload_json 类型无效');
}

let extra = parseExtraPayload(getVar('jwt_payload_json'));

const payload = Object.assign({}, extra, {
    account: account,
    iat: now,
    exp: now + ttl,
});

const header = { alg: 'HS256', typ: 'JWT' };
const headerB64 = base64urlEncodeUtf8(JSON.stringify(header));
const payloadB64 = base64urlEncodeUtf8(JSON.stringify(payload));
const unsigned = headerB64 + '.' + payloadB64;
const signatureB64 = hmacSha256Base64Url(unsigned, secret);
const token = unsigned + '.' + signatureB64;

const bodyJson = JSON.stringify({ jwt: token });
// 将请求 Body 设为 raw（建议 JSON），预请求会写入完整 body
pm.request.body.raw = bodyJson;
