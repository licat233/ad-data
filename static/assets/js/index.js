new Vue({
    el: "#app",
    data() {
        this.chartSettings = {
            axisSite: {right: ['CV', 'CPA']},
            yAxisType: ['normal', 'normal'],
            yAxisName: ['CNY', 'CV/CPA'],
            series: {
                label: {
                    normal: {
                        show: true
                    }
                }
            },
            labelMap: {
                Date: '时间',
                CNY: '消耗',
                CV: '加粉',
                CPA: '成本'
            }
        };
        return {
            chartData: {
                columns: ['Date', 'CNY', 'CV', 'CPA'],
                rows: []
            },
            DataShow: [],
            onedate: '',
            rangedate: '',
            ckprojectname: '',
            ckprojectid: '',
            playproject:'',
            todaySpend: 0,
            todayFans: 0,
            todayCpa: 0,
            campaigns: [],
            lines: {},
            projectStatus: false,
            projectlist: [],
            resalldata: [],
            Timingname: "停止",
            Timing: true,
            pickerOptions: {
                disabledDate(time) {
                    return time.getTime() > Date.now()
                },
                shortcuts: [
                    {
                        text: '今天',
                        onClick(picker) {
                            picker.$emit('pick', new Date())
                        }
                    },
                    {
                        text: '昨天',
                        onClick(picker) {
                            const date = new Date()
                            date.setTime(date.getTime() - 3600 * 1000 * 24)
                            picker.$emit('pick', date)
                        }
                    },
                    {
                        text: '一周前',
                        onClick(picker) {
                            const date = new Date()
                            date.setTime(date.getTime() - 3600 * 1000 * 24 * 7)
                            picker.$emit('pick', date)
                        }
                    }
                ]
            },
            newprojectform: {
                projectname: '',
                AccountID: '',
                CampaignId: '',
                linename: '',
                lineid: '',
            },
            showform: false,
            form1FormRules: {
                projectname: [
                    {required: true, message: '请输入项目名称', trigger: 'blur'}
                ],
                AccountID: [
                    {required: true, message: '请输入AccountID', trigger: 'blur'}
                ],
                CampaignId: [
                    {required: true, message: '请输入CampaignID', trigger: 'blur'}
                ],
                linename: [
                    {required: true, message: '请输入linename', trigger: 'blur'}
                ],
                lineid: [
                    {required: true, message: '请输入lineid', trigger: 'blur'}
                ]
            },
        }
    },
    async mounted() {
        let resall = await axios.post('/api/GetAllPJdata')
        this.resalldata = resall.data.jsondata
        this.Timing = !resall.data.timingstatus
        // console.log(resall)
        for (let i = 0; i < this.resalldata.length; i++) {
            this.projectlist.push({
                "id": this.resalldata[i].PId,
                "name": this.resalldata[i].Name
            })
            if (this.resalldata[i].Status === 1){
                this.ckprojectid = this.resalldata[i].PId
            }
        }

    },
    methods: {
        logout() {
            window.sessionStorage.clear()
            location.href = '/login'
        },
        handleOpen(key, keyPath) {
            console.log(key, keyPath)
        },
        handleClose(key, keyPath) {
            console.log(key, keyPath)
        },
        mstatus() {
            this.isCollapse = !this.isCollapse
            this.isCollapse ? (this.Menuname = '打开') : (this.Menuname = '收起')
        },
        getPdatabyCK(id) {
            let resu = axios.post('/api/PJdatabyID', {"projectid": id});
            console.log(resu)
            return resu
        },
        formcancel() {
            this.showform = false
            this.$refs.form1.resetFields()
        },
        addproject() {
            this.$refs.form1.validate(
                async valid => {
                    if (!valid) return false
                    const {data: res} = await axios.post('/api/addproject', this.newprojectform)
                    // console.log(res)
                    if (res.status !== 200) {
                        this.$message.error('添加失败！')
                    } else {
                        this.$message.success('添加成功！')
                        this.showform = false
                        this.$refs.form1.resetFields()
                    }
                }
            )
        },
        async mdfProjectStatus() {
            // console.log("状态改变了")
            let statuscode = this.projectStatus === true ? "1" : "0";
            const res = await axios.post('/api/modifyprojectStatus', {
                "projectid": this.ckprojectid,
                "statuscode": statuscode
            })
            // console.log(res)
            let msg = this.projectStatus === true ? "开启" : "关闭";
            if (res.status !== 200) {
                this.$message.error(msg + '失败！')
                return false
            }
            if (this.projectStatus) {
                this.$message.success('已' + msg)
            } else {
                this.$message.warning('已' + msg)
            }

        },
        async TimingF() {
            instruct = this.Timing === true ? "on":"off";
            const res = await axios.post('/api/instructTiming', {"instruct":instruct})
            if (res.status === 200 && this.Timing === true){
                this.Timing = !this.Timing;
                this.$message.success('开始监控项目数据');
            }else if(res.status === 200 && this.Timing !== true){
                this.Timing = !this.Timing;
                this.$message.success('关闭监控项目数据');
            }else{
                this.$message.warning('失败！');
            }
        }
    },
    watch: {
        ckprojectid: async function (newvalue, oldvalue) {
            // console.log('new:%s,old:%s',newvalue,oldvalue)
            let resu = await axios.post('/api/PJdatabyID', {"projectid": newvalue});
            // console.log(resu.data.jsondata.Campaignid)
            this.projectStatus = resu.data.jsondata.Status === 1;
            this.campaigns = [];
            this.lines = [];
            let campaignsY = resu.data.jsondata.Campaignid;
            for (let i = 0; i < campaignsY.length; i++) {
                this.campaigns.push({
                    "campaignid": campaignsY[i]
                })
            }
            this.lines = resu.data.jsondata.Lineid
            let res = await axios.post('/api/todaydata', {"projectid": newvalue});
            // console.log(res)
            if (res.data.jsondata === null || res.data.jsondata === undefined) {
                this.todaySpend = "无数据";
                this.todayFans = "无数据";
                this.todayCpa = "无数据";
                this.chartData.rows = [];
                return false
            }
            this.todaySpend = res.data.jsondata.Spend;
            this.todayFans = res.data.jsondata.Fans;
            this.todayCpa = res.data.jsondata.Cpa;
            // console.log(res)
            let t;
            let res1 = await axios.post('/api/projectdata', {"projectid": newvalue});
            if (res1.data.jsondata === null || res1.data.jsondata === undefined) {
                this.chartData.rows = [];
                return false
            }
            this.DataShow = res1.data.jsondata;
            for (let i = 0; i < this.DataShow.length; i++) {
                t = this.DataShow[i].Date;
                this.chartData.rows.push({
                    "Date": FormatDate(t),
                    "CNY": this.DataShow[i].CNY,
                    "CV": this.DataShow[i].CV,
                    "CPA": this.DataShow[i].CPA
                })
            }
        }
    },
    components: {}
})

/**
 * @return {string}
 */
function FormatDate(t) {
    date = new Date(t * 1000)
    let year = date.getFullYear(),
        month = ("0" + (date.getMonth() + 1)).slice(-2),
        sdate = ("0" + date.getDate()).slice(-2),
        hour = ("0" + date.getHours()).slice(-2),
        minute = ("0" + date.getMinutes()).slice(-2),
        second = ("0" + date.getSeconds()).slice(-2);
    // 拼接返回
    return month + "-" + sdate + " " + hour + ":" + minute;
}