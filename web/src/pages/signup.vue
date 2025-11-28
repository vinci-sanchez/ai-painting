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
            <div class="form-footer">
              <el-checkbox class="form-agreement" v-model="accountForm.agree">
                我已阅读并同意《用户使用手册》
              </el-checkbox>
              <el-button
                type="primary"
                :loading="accountLoading"
                class="form-submit"
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
              <el-checkbox class="form-agreement" v-model="phoneForm.agree">
                我已阅读并同意《用户使用手册》
              </el-checkbox>
              <el-button
                type="primary"
                :loading="phoneLoading"
                class="form-submit"
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

const BACK_URL =
  (config as Record<string, string | undefined>).BACK_URL ??
  "http://localhost:3000";

const activeTab = ref<"account" | "phone">("account");

const accountForm = reactive({
  username: "",
  password: "",
  confirmPassword: "",
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

const serverSmsCode = ref("");
const smsCountdown = ref(0);
const smsLoading = ref(false);
const accountLoading = ref(false);
const phoneLoading = ref(false);
let countdownTimer: number | null = null;

const parseJsonSafe = async (response: Response) => {
  const raw = await response.text();
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    throw new Error(raw.trim() || "服务器返回了非 JSON 数据");
  }
};

const validateConfirmPassword = (
  _rule: unknown,
  value: string,
  callback: (err?: Error) => void
) => {
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

const generateSmsCode = (length = 6) => {
  return Array.from({ length }, () => Math.floor(Math.random() * 10)).join("");
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

const fetchSmsCode = () => {
  if (!/^1[3-9]\d{9}$/.test(phoneForm.phone)) {
    ElMessage.warning("请先输入合法的手机号");
    return;
  }
  smsLoading.value = true;
  setTimeout(() => {
    serverSmsCode.value = generateSmsCode();
    smsLoading.value = false;
    ElMessage.success(`验证码已发送：${serverSmsCode.value}`);
    startSmsCountdown();
  }, 300);
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
  accountLoading.value = true;
  try {
    const response = await fetch(`${BACK_URL}/api/users`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: accountForm.username,
        password: accountForm.password,
      }),
    });
    const data = await parseJsonSafe(response);
    if (!response.ok) {
      throw new Error(data.error || data.message || "注册失败");
    }
    ElMessage.success("注册成功，请登录");
    router.push("/login");
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
    const response = await fetch(`${BACK_URL}/api/users`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: phoneForm.phone,
        password: phoneForm.password,
      }),
    });
    const data = await parseJsonSafe(response);
    if (!response.ok) {
      throw new Error(data.error || data.message || "注册失败");
    }
    ElMessage.success("注册成功，请登录");
    router.push("/login");
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
  width: min(460px, 100%);
  border-radius: 24px;
  box-shadow: 0 18px 40px rgba(31, 42, 68, 0.12);
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
  gap: 12px;
  align-items: flex-start;
  margin-top: 12px;
  flex-wrap: wrap;
}

.form-agreement {
  flex: 1;
  line-height: 1.4;
  color: #4b5563;
}

.form-submit {
  min-width: 148px;
}

.inline-tag {
  margin-top: 8px;
}

.signup-switch {
  text-align: center;
  margin-top: 12px;
  color: #7c8db5;
}

@media (max-width: 640px) {
  .signup-page {
    padding: 16px 12px 24px;
  }

  .signup-card {
    border-radius: 18px;
    box-shadow: 0 12px 30px rgba(31, 42, 68, 0.18);
  }

  .signup-header h2 {
    font-size: 20px;
  }

  .signup-header p {
    font-size: 13px;
  }

  .signup-tabs :deep(.el-tabs__header) {
    margin-bottom: 16px;
  }

  .form-footer {
    flex-direction: column;
    align-items: stretch;
  }

  .form-agreement {
    width: 100%;
  }

  .form-submit {
    width: 100%;
  }
}

</style>
