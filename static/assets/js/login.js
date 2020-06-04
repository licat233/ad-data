new Vue({
    el: "#login",
    data: {
        // 这是登陆表单的数据绑定对象
        loginForm: {
            username: 'admin',
            password: '123456'
        },
        // 这是表单的验证规则
        loginFormRules: {
            // 验证用户名是否合法
            username: [
                { required: true, message: '请输入账号', trigger: 'blur' },
                { min: 3, max: 10, message: '长度在 3 到 10 个字符', trigger: 'blur' }
            ],
            // 验证密码是否合法
            password: [
                { required: true, message: '请输入密码', trigger: 'blur' },
                { min: 6, max: 10, message: '长度在 6 到 12 个字符', trigger: 'blur' }
            ]
        }
    },
    methods: {
        submitLoginForm() {
            this.$refs.loginFormRef.validate(
                async valid => {
                    if (!valid) return false
                    const {data: res } = await axios.post('/api/login', this.loginForm)
                    // console.log(res)
                    if (res.status !== 200) {
                        return this.$message.error('登录失败！')
                    }else{
                        this.$message.success('登录成功！')
                        location.href='/'
                    }
                }
            )
        },
        resetLoginForm() {
            this.$refs.form1.resetFields()
        }
    }
})
