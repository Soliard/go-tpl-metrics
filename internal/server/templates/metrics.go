// Package templates содержит шаблон для вывода html документа клиенту
package templates

// MetricsTemplate это верстка страницы для GET запроса с полным перечнем метрик
const MetricsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
    <style>
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
    </style>
</head>
<body>
    <h1>Metrics</h1>
    <table>
        <tr>
            <th>ID</th>
            <th>Type</th>
            <th>Value</th>
            <th>Delta</th>
            <th>Hash</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.MType}}</td>
            <td>{{.StringifyValue}}</td>
            <td>{{.StringifyDelta}}</td>
            <td>{{.Hash}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`
