// ================== 飞行架次统计 ================== //

// 直接定义 option1，去除 data/myData 依赖
option1 = {
    title: {
        show: true,
        text: '',
        subtext: '',
        link: ''
    },
    // tooltip: {
    //     trigger: 'axis',
    //     axisPointer: { type: 'none' },
    //     formatter: function(params) {
    //         var time = '';
    //         var str = '';
    //         for (var i of params) {
    //             time = i.name.replace(/\n/g, '') + '<br/>';
    //             if (i.data == 'null' || i.data == null) {
    //                 str += i.seriesName + '：无数据' + '<br/>'
    //             } else {
    //                 str += i.seriesName + '：' + i.data + '<br/>'
    //             }
    //         }
    //         return time + str;
    //     }
    // },
    legend: {
        right: 10,
        top: 0,
        itemGap: 16,
        itemWidth: 10,
        itemHeight: 10,
        data: [],
        textStyle: {
            color: '#fff',
            fontStyle: 'normal',
            fontFamily: '微软雅黑',
            fontSize: 12,
        }
    },
    grid: {
        x: 0,
        y: 40,
        x2: 0,
        y2: 40,
    },
    xAxis: {
        type: 'category',
        data: [],
        axisTick: { show: false },
        axisLine: { show: false },
        axisLabel: {
            show: true,
            interval: 0,
            textStyle: {
                lineHeight: 5,
                padding: [2, 2, 0, 2],
                height: 50,
                fontSize: 12,
                color: '#fff',
            },
            rich: {
                Sunny: {
                    height: 50,
                    padding: [0, 5, 0, 5],
                    align: 'center',
                },
            },
            formatter: function(params) {
                // 保持原有多行显示逻辑
                // var newParamsName = "";
                // var splitNumber = 5;
                // var paramsNameNumber = params && params.length;
                // if (paramsNameNumber && paramsNameNumber <= 4) {
                //     splitNumber = 4;
                // } else if (paramsNameNumber >= 5 && paramsNameNumber <= 7) {
                //     splitNumber = 4;
                // } else if (paramsNameNumber >= 8 && paramsNameNumber <= 9) {
                //     splitNumber = 5;
                // } else if (paramsNameNumber >= 10 && paramsNameNumber <= 14) {
                //     splitNumber = 5;
                // } else {
                //     params = params && params.slice(0, 15);
                // }
                // var provideNumber = splitNumber;
                // var rowNumber = Math.ceil(paramsNameNumber / provideNumber) || 0;
                // if (paramsNameNumber > provideNumber) {
                //     for (var p = 0; p < rowNumber; p++) {
                //         var tempStr = "";
                //         var start = p * provideNumber;
                //         var end = start + provideNumber;
                //         if (p == rowNumber - 1) {
                //             tempStr = params.substring(start, paramsNameNumber);
                //         } else {
                //             tempStr = params.substring(start, end) + "\n";
                //         }
                //         newParamsName += tempStr;
                //     }
                // } else {
                //     newParamsName = params;
                // }
                // params = newParamsName
                return '{Sunny|' + params + '}';
            },
            color: '#687284',
        },
    },
    yAxis: {
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { show: false },
        splitLine: {
            show: true,
            lineStyle: {
                color: '#F1F3F5',
                type: 'solid'
            },
            interval: 2
        },
        splitNumber: 4,
    },
    series: [{
        name: '',
        type: 'bar',
        barGap: '0.5px',
        data: [],
        barWidth: 12,
        label: {
            normal: {
                show: true,
                formatter: '{c}',
                position: 'top',
                textStyle: {
                    color: '#fff',
                    fontStyle: 'normal',
                    fontFamily: '微软雅黑',
                    textAlign: 'center',
                    fontSize: 14,
                },
            },
        },
        itemStyle: {
            normal: {
                barBorderRadius: 0,
                borderWidth: 1,
                borderColor: '#ddd',
                color: '#009883'
            },
        }
    }]
}
// ================== 飞行架次统计 end ============== //

//////////////////////交通工具流量
option2 = {
    
    tooltip: {//鼠标指上时的标线
        trigger: 'axis',
        axisPointer: {
            lineStyle: {
                color: '#fff'
            }
        }
    },
    legend: {
        icon: 'rect',
        itemWidth: 14,
        itemHeight: 5,
        itemGap: 13,
        data: ['小型车', '中型车', '大型车'],
        right: '10px',
        top: '0px',
        textStyle: {
            fontSize: 12,
            color: '#fff'
        }
    },
    grid: {
        x: 35,
        y: 25,
        x2: 8,
        y2: 25,
    },
    xAxis: [{
        type: 'category',
        boundaryGap: false,
        axisLine: {
            lineStyle: {
                color: '#57617B'
            }
        },
        axisLabel: {
            textStyle: {
                color:'#fff',
            },
        },
        data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
    }],
    yAxis: [{
        type: 'value',
        axisTick: {
            show: false
        },
        axisLine: {
            lineStyle: {
                color: '#57617B'
            }
        },
        axisLabel: {
            margin: 10,
            textStyle: {
                fontSize: 14
            },
            textStyle: {
                color:'#fff',
            },
        },
        splitLine: {
            lineStyle: {
                color: '#57617B'
            }
        }
    }],
    series: [{
        name: '小型车',
        type: 'line',
        smooth: true,
        lineStyle: {
            normal: {
                width: 2
            }
        },
        areaStyle: {
            normal: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
                    offset: 0,
                    color: 'rgba(137, 189, 27, 0.3)'
                }, {
                    offset: 0.8,
                    color: 'rgba(137, 189, 27, 0)'
                }], false),
                shadowColor: 'rgba(0, 0, 0, 0.1)',
                shadowBlur: 10
            }
        },
        itemStyle: {
            normal: {
                color: 'rgb(137,189,27)'
            }
        },
        data: [20,35,34,45,52,41,49,64,24,52.4,24,33]
    }, {
        name: '中型车',
        type: 'line',
        smooth: true,
        lineStyle: {
            normal: {
                width: 2
            }
        },
        areaStyle: {
            normal: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
                    offset: 0,
                    color: 'rgba(0, 136, 212, 0.3)'
                }, {
                    offset: 0.8,
                    color: 'rgba(0, 136, 212, 0)'
                }], false),
                shadowColor: 'rgba(0, 0, 0, 0.1)',
                shadowBlur: 10
            }
        },
        itemStyle: {
            normal: {
                color: 'rgb(0,136,212)'
            }
        },
        data: [97.3,99.2,99.3,100.0,99.6,90.6,80.0,91.5,69.8,67.5,90.4,84.9]
    }, {
        name: '大型车',
        type: 'line',
        smooth: true,
        lineStyle: {
            normal: {
                width: 2
            }
        },
        areaStyle: {
            normal: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
                    offset: 0,
                    color: 'rgba(219, 50, 51, 0.3)'
                }, {
                    offset: 0.8,
                    color: 'rgba(219, 50, 51, 0)'
                }], false),
                shadowColor: 'rgba(0, 0, 0, 0.1)',
                shadowBlur: 10
            }
        },
        itemStyle: {
            normal: {
                color: 'rgb(219,50,51)'
            }
        },
        data: [84.2,81.0,67.5,62.1,43.7,68.5,51.9,71.8,76.7,67.6,62.9,0]
    }, ]
};
//////////////////////交通工具流量 end

//////////////////////本月发生事件1
var color = ['#e9df3d', '#f79c19', '#21fcd6', '#08c8ff', '#df4131'];
var data = [{
        "name": "xxx",
        "value": 30
    },
    {
        "name": "xxx",
        "value": 30
    },
    {
        "name": "xxx",
        "value": 42
    },
    {
        "name": "xxx",
        "value": 50
    },
    {
        "name": "xxx",
        "value": 34
    }
];

var max = data[0].value;
data.forEach(function(d) {
    max = d.value > max ? d.value : max;
});

var renderData = [{
    value: [],
    name: "告警类型TOP5",
    symbol: 'none',
    lineStyle: {
        normal: {
            color: '#ecc03e',
            width: 2
        }
    },
    areaStyle: {
        normal: {
            color: new echarts.graphic.LinearGradient(0, 0, 1, 0,
                [{
                    offset: 0,
                    color: 'rgba(203, 158, 24, 0.8)'
                }, {
                    offset: 1,
                    color: 'rgba(190, 96, 20, 0.8)'
                }],
                false)
        }
    }
}];


data.forEach(function(d, i) {
    var value = ['', '', '', '', ''];
    value[i] = max,
    renderData[0].value[i] = d.value;
    renderData.push({
        value: value,
        symbol: 'circle',
        symbolSize: 12,
        lineStyle: {
            normal: {
                color: 'transparent'
            }
        },
        itemStyle: {
            normal: {
                color: color[i],
            }
        }
    })
})
var indicator = [];

data.forEach(function(d) {
    indicator.push({
        name: d.name,
        max: max,
        color: '#fff'
    })
})


option3 = {
    tooltip: {
        show: true,
        trigger: "item"
    },
    radar: {
        center: ["50%", "50%"],//偏移位置
        radius: "80%",
        startAngle: 40, // 起始角度
        splitNumber: 4,
        shape: "circle",
        splitArea: {
            areaStyle: {
                color: 'transparent'
            }
        },
        axisLabel: {
            show: false,
            fontSize: 20,
            color: "#000",
            fontStyle: "normal",
            fontWeight: "normal"
        },
        axisLine: {
            show: true,
            lineStyle: {
                color: "rgba(255, 255, 255, 0.5)"
            }
        },
        splitLine: {
            show: true,
            lineStyle: {
                color: "rgba(255, 255, 255, 0.5)"
            }
        },
        indicator: indicator
    },
    series: [{
        type: "radar",
        data: renderData
    }]
}
//////////////////////本月发生事件1 end
