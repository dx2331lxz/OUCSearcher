<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Custom Search Page</title>
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        .search-bar {
            margin: 10px 0;
        }
        .results {
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container-fluid">
        <div class="row">
            <!-- 搜索框 -->
            <div class="col-12">
                <form method="GET" action="/search" class="d-flex align-items-center search-bar" id="searchForm">
                    <input
                        type="text"
                        name="q"
                        class="form-control w-25"
                        placeholder="Search..."
                        value="{{ .Query }}"
                        onkeydown="if (event.key === 'Enter') document.getElementById('searchForm').submit();"
                    >
                    <button type="submit" class="btn btn-primary ms-2">Search</button>
                </form>
            </div>
        </div>
        <div class="row">
            <!-- 搜索结果 -->
            <div class="col-md-8">
                <div class="results">
                    {{ range .Results }}
                        <div class="card mb-3">
                            <div class="card-body">
                                <h5 class="card-title"><a href="{{ .URL }}" target="_blank">{{ .Title }}</a></h5>
                                <p class="card-text">{{ .Description }}</p>
                            </div>
                        </div>
                    {{ else }}
                        <p>No results found.</p>
                    {{ end }}
                </div>
            </div>
            <!-- 右侧内容 -->
            <div class="col-md-4">
                <div class="p-3 bg-light border rounded">
                    <h4>右侧内容</h4>
                    <p>这里可以放置广告、链接或其他信息。</p>
                </div>
            </div>
        </div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>