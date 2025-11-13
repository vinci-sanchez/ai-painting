<template>
  <div class="signup-page">
    <el-card class="signup-card" shadow="hover">
      <div class="signup-header">
        <h2>Text-to-manga Studio</h2>
        <p>文生图 · 新用户注册</p>
      </div>

      <el-tabs v-model="activeTab" stretch class="signup-tabs">
        <el-tab-pane label="账号注册" name="account">
          <el-form
            ref="accountFormRef"
            :model="accountForm"
            :rules="accountRules"
            label-position="top"
            class="signup-form"
          >
            <el-form-item label="用户名" prop="username">
              <el-input
                v-model="accountForm.username"
                placeholder="请输入用户名"
                maxlength="10"
              />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="accountForm.password"
                placeholder="请输入密码，8位字母或数字组合"
                maxlength="20"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model="accountForm.confirmPassword"
                placeholder="请再次输入密码"
                maxlength="20"
                show-password
              />
            </el-form-item>
            <el-form-item label="验证码" prop="captcha">
              <el-input
                v-model="accountForm.captcha"
                placeholder="请输入图形验证码"
                maxlength="8"
              >
                <template #suffix>
                  <el-button
                    link
                    type="primary"
                    :loading="captchaLoading"
                    @click="fetchCaptcha"
                  >
                    获取验证码
                  </el-button>
                </template>
              </el-input>
              <el-tag
                v-if="captchaValue"
                type="success"
                size="small"
                class="inline-tag"
              >
                {{ captchaValue }}
              </el-tag>
            </el-form-item>
            <div class="form-footer">
              <el-checkbox v-model="accountForm.agree">
                我已阅读并同意《用户使用手册》
              </el-checkbox>
              <el-button
                type="primary"
                :loading="accountLoading"
                @click="handleAccountSignup"
              >
                注册
              </el-button>
            </div>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="手机号注册" name="phone">
          <el-form
            ref="phoneFormRef"
            :model="phoneForm"
            :rules="phoneRules"
            label-position="top"
            class="signup-form"
          >
            <el-form-item label="手机号" prop="phone">
              <el-input
                v-model="phoneForm.phone"
                placeholder="请输入手机号"
                maxlength="11"
              />
            </el-form-item>
            <el-form-item label="登录密码" prop="password">
              <el-input
                v-model="phoneForm.password"
                placeholder="设置 8 位字母或数字组合密码"
                maxlength="20"
                show-password
              />
            </el-form-item>
            <el-form-item label="短信验证码" prop="code">
              <el-input
                v-model="phoneForm.code"
                placeholder="请输入短信验证码"
                maxlength="6"
              >
                <template #suffix>
                  <el-button
                    link
                    type="primary"
                    :disabled="smsCountdown > 0"
                    :loading="smsLoading"
                    @click="fetchSmsCode"
                  >
                    {{ smsCountdown > 0 ? `${smsCountdown}s 后重试` : "获取验证码" }}
                  </el-button>
                </template>
              </el-input>
            </el-form-item>
            <div class="form-footer">
              <el-checkbox v-model="phoneForm.agree">
                我已阅读并同意《用户使用手册》
              </el-checkbox>
              <el-button
                type="primary"
                :loading="phoneLoading"
                @click="handlePhoneSignup"
              >
                注册
              </el-button>
            </div>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="signup-switch">
        已有账号？
        <el-button link type="primary" @click="goLogin">
          返回登录
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
defineOptions({ name: "signup" });

import { onBeforeUnmount, reactive, ref } from "vue";
import type { FormInstance, FormRules } from "element-plus";
import { ElMessage } from "element-plus";
import router from "../router.js";
import config from "./config.json";

type ApiResponse<T = Record<string, unknown>> = {
  status?: number;
  message?: string;
  token?: string;
  code?: string;
  mark_num?: string;
  data?: T;
};

const AUTH_BASE_URL =
  (config as Record<string, string | undefined>).AUTH_BASE_URL ??
  "http://127.0.0.1:8900";

const activeTab = ref<"account" | "phone">("account");

const accountForm = reactive({
  username: "",
  password: "",
  confirmPassword: "",
  captcha: "",
  agree: false,
});

const phoneForm = reactive({
  phone: "",
  password: "",
  code: "",
  agree: false,
});

const accountFormRef = ref<FormInstance>();
const phoneFormRef = ref<FormInstance>();

const captchaValue = ref("");
const serverSmsCode = ref("");
const smsCountdown = ref(0);
const captchaLoading = ref(false);
const smsLoading = ref(false);
const accountLoading = ref(false);
const phoneLoading = ref(false);
let countdownTimer: number | null = null;

const validateConfirmPassword = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (!value) {
    callback(new Error("请再次输入密码"));
  } else if (value !== accountForm.password) {
    callback(new Error("两次输入的密码不一致"));
  } else {
    callback();
  }
};

const accountRules: FormRules = {
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    { min: 2, max: 10, message: "用户名长度为 2-10 位", trigger: "blur" },
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    {
      pattern: /^[0-9a-zA-Z]{8}$/,
      message: "密码需为 8 位字母或数字组合",
      trigger: "blur",
    },
  ],
  confirmPassword: [{ validator: validateConfirmPassword, trigger: "blur" }],
  captcha: [{ required: true, message: "请输入验证码", trigger: "blur" }],
};

const phoneRules: FormRules = {
  phone: [
    { required: true, message: "请输入手机号", trigger: "blur" },
    {
      pattern: /^1[3-9]\d{9}$/,
      message: "请输入合法的 11 位手机号",
      trigger: "blur",
    },
  ],
  password: [
    { required: true, message: "请设置密码", trigger: "blur" },
    {
      pattern: /^[0-9a-zA-Z]{8}$/,
      message: "密码需为 8 位字母或数字组合",
      trigger: "blur",
    },
  ],
  code: [{ required: true, message: "请输入验证码", trigger: "blur" }],
};

const handleAgreementCheck = (agree: boolean) => {
  if (!agree) {
    ElMessage.warning("请先阅读并同意《用户使用手册》");
    return false;
  }
  return true;
};

const fetchCaptcha = async () => {
  captchaLoading.value = true;
  captchaValue.value = "";
  try {
    const response = await fetch(`${AUTH_BASE_URL}/get_code`);
    const data = (await response.json()) as ApiResponse;
    if (response.ok && data.status === 200 && data.code) {
      captchaValue.value = data.code;
      ElMessage.success("验证码已生成");
    } else {
      throw new Error(data.message || "获取验证码失败");
    }
  } catch (error) {
    const err = error as Error;
    ElMessage.error(err.message || "验证码获取异常");
  } finally {
    captchaLoading.value = false;
  }
};

const startSmsCountdown = () => {
  smsCountdown.value = 60;
  if (countdownTimer) {
    clearInterval(countdownTimer);
  }
  countdownTimer = window.setInterval(() => {
    if (smsCountdown.value <= 1) {
      smsCountdown.value = 0;
      if (countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
      }
    } else {
      smsCountdown.value -= 1;
    }
  }, 1000);
};

const fetchSmsCode = async () => {
  if (!/^1[3-9]\d{9}$/.test(phoneForm.phone)) {
    ElMessage.warning("请先输入合法的手机号");
    return;
  }
  smsLoading.value = true;
  serverSmsCode.value = "";
  try {
    const response = await fetch(`${AUTH_BASE_URL}/get_tell_mark`);
    const data = (await response.json()) as ApiResponse;
    if (response.ok && data.status === 200 && data.mark_num) {
      serverSmsCode.value = data.mark_num;
      ElMessage.success("验证码已发送，请留意短信");
      startSmsCountdown();
    } else {
      throw new Error(data.message || "验证码获取失败");
    }
  } catch (error) {
    const err = error as Error;
    ElMessage.error(err.message || "验证码获取异常");
  } finally {
    smsLoading.value = false;
  }
};

const handleAccountSignup = async () => {
  try {
    await accountFormRef.value?.validate();
  } catch {
    return;
  }
  if (!handleAgreementCheck(accountForm.agree)) {
    return;
  }
  if (!captchaValue.value || accountForm.captcha !== captchaValue.value) {
    ElMessage.error("验证码错误，请重新获取");
    return;
  }
  accountLoading.value = true;
  try {
    const response = await fetch(`${AUTH_BASE_URL}/register_pwd`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: accountForm.username,
        password: accountForm.password,
        mark2: accountForm.captcha,
      }),
    });
    const data = (await response.json()) as ApiResponse;
    if (response.ok && data.status === 200) {
      ElMessage.success("注册成功，请登录");
      router.push("/login");
    } else {
      throw new Error(data.message || "注册失败");
    }
  } catch (error) {
    const err = error as Error;
    ElMessage.error(err.message || "注册失败，请稍后重试");
  } finally {
    accountLoading.value = false;
  }
};

const handlePhoneSignup = async () => {
  try {
    await phoneFormRef.value?.validate();
  } catch {
    return;
  }
  if (!handleAgreementCheck(phoneForm.agree)) {
    return;
  }
  if (!serverSmsCode.value || phoneForm.code !== serverSmsCode.value) {
    ElMessage.error("验证码错误，请重新获取");
    return;
  }
  phoneLoading.value = true;
  try {
    const response = await fetch(`${AUTH_BASE_URL}/register_tell`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tell: phoneForm.phone,
        password: phoneForm.password,
        mark1: phoneForm.code,
      }),
    });
    const data = (await response.json()) as ApiResponse;
    if (response.ok && data.status === 200) {
      ElMessage.success("注册成功，请登录");
      router.push("/login");
    } else {
      throw new Error(data.message || "注册失败");
    }
  } catch (error) {
    const err = error as Error;
    ElMessage.error(err.message || "注册失败，请稍后重试");
  } finally {
    phoneLoading.value = false;
  }
};

const goLogin = () => {
  router.push("/login");
};

onBeforeUnmount(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
});
</script>

<style scoped>
.signup-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  background-color: transparent;
}

.signup-card {
  width: 460px;
  max-width: 100%;
  border-radius: 24px;
}

.signup-header {
  text-align: center;
  margin-bottom: 16px;
}

.signup-header h2 {
  margin: 0;
  font-size: 24px;
  color: #1f2a44;
}

.signup-header p {
  margin: 4px 0 0;
  color: #7c8db5;
  font-size: 14px;
}

.signup-tabs {
  --el-border-color-light: rgba(117, 140, 255, 0.2);
}

.signup-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: #4b5563;
}

.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.inline-tag {
  margin-top: 8px;
}

.signup-switch {
  text-align: center;
  margin-top: 12px;
  color: #7c8db5;
}
</style>
