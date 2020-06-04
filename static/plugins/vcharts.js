new Vue({
    el: '#myvcharts',
    props: ['ckprojectid'],
    data() {
        this.chartSettings = {
            axisSite: { right: ['CV', 'CPA'] },
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
                columns: ['Date', 'CNY','CV','CPA'],
                rows: []
                // rows: [
                //     { Date: '1/1', CNY: 1393, CV: 13, CPA: 110 },
                //     { Date: '1/2', CNY: 3530, CV: 35, CPA: 90 },
                //     { Date: '1/3', CNY: 2923, CV: 39, CPA: 85 },
                //     { Date: '1/4', CNY: 1723, CV: 17, CPA: 101 },
                //     { Date: '1/5', CNY: 3792, CV: 38, CPA: 99 },
                //     { Date: '1/6', CNY: 4593, CV: 46, CPA: 95 }
                // ],
            },
            DataShow:[]
        }
    },
    watch: {
        //正确给 ckprojectid 赋值的 方法
        ckprojectid:function(newVal){
            console.log(newVal)
            //TODO 卡在props传值
        }
    },
    // async mounted() {
    //     console.log(this.CKPrid)
    //     let t;
    //     let res = await axios.post('/api/projectdata', {"projectid": this.ckprojectid});
    //     this.DataShow = res.data.jsondata;
    //     for (let i=0; i<this.DataShow.length; i++) {
    //         t = this.DataShow[i].Date;
    //         this.chartData.rows.push({
    //             "Date": FormatDate(t),
    //             "CNY": this.DataShow[i].CNY,
    //             "CV": this.DataShow[i].CV,
    //             "CPA": this.DataShow[i].CPA
    //         })
    //     }
    //     // console.log(this.chartData.rows);
    //     // console.log(this.DataShow);
    // },
    method: {
    }
});
/**
 * @return {string}
 */
function FormatDate(t){
    date = new Date(t*1000)
    let year = date.getFullYear(),
        month = ("0" + (date.getMonth() + 1)).slice(-2),
        sdate = ("0" + date.getDate()).slice(-2),
        hour = ("0" + date.getHours()).slice(-2),
        minute = ("0" + date.getMinutes()).slice(-2),
        second = ("0" + date.getSeconds()).slice(-2);
    // 拼接返回
    return month + "-" + sdate + " " + hour + ":" + minute ;
}