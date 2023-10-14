{
    title: '{{ Title }}',
    uid: '{{ UID }}',
    datasource: {
        type: "prometheus",
        uid: "${cluster}"
    },
    gridPos: [
        {
            h: 1, w: 24, x: 0, y: {{ calculateYByHeight 1 }},
        },
        {
            h: 9, w: 24, x: 0, y: {{ calculateYByHeight 9 }},
        },
        {
             h: 10, w: 24, x: 0, y: {{ calculateYByHeight 10 }},
        },
        {
            h: 8, w: 24, x: 0, y: {{ calculateYByHeight 8 }},
        },
        {
            h: 8, w: 24, x: 0, y: {{ calculateYByHeight 8 }},
        },
        {
            h: 8, w: 24, x: 0, y: {{ calculateYByHeight 8 }},
        },
        {
            h: 10, w: 12, x: 0, y: {{ calculateYByHeight 10 }},
        },
        {
            h: 10, w: 12, x: 12, y: {{ calculateYByHeight 10 }},
        },
        {
            h: 8, w: 24, x: 0, y: {{ calculateYByHeight 8 }},
        },
        {
            h: 9, w: 24, x: 0, y: {{ calculateYByHeight 9 }},   // Requests per seconds by methods panel
        }
    ],
}