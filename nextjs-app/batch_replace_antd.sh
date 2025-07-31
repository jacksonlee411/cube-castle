#!/bin/bash

# 批量替换 antd 引用脚本

# 替换 import 语句
find /home/shangmeilin/cube-castle/nextjs-app/src -name "*.tsx" -o -name "*.ts" | xargs sed -i "
s|import { notification } from 'antd'|import { toast } from 'react-hot-toast'|g
s|import { Alert } from 'antd'|// Alert replaced with custom alert component|g
s|import { Button } from 'antd'|import { Button } from '@/components/ui/button'|g
s|import { Space } from 'antd'|// Space replaced with div + className=\"flex gap-2\"|g
s|import { Card } from 'antd'|import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'|g
s|import { Input } from 'antd'|import { Input } from '@/components/ui/input'|g
s|import { Select } from 'antd'|import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'|g
s|import { Modal } from 'antd'|import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'|g
s|import { Table } from 'antd'|import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'|g
s|import { Form } from 'antd'|import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'|g
s|import { DatePicker } from 'antd'|import { DatePicker } from '@/components/ui/date-picker'|g
s|import { Spin } from 'antd'|import { Spin } from '@/components/ui/spin'|g
s|import { Typography } from 'antd'|import { Typography, Title, Text } from '@/components/ui/typography'|g
s|import { Divider } from 'antd'|import { Divider } from '@/components/ui/divider'|g
s|import { Progress } from 'antd'|import { Progress } from '@/components/ui/progress'|g
s|import { Timeline } from 'antd'|import { Timeline, TimelineItem } from '@/components/ui/timeline'|g
s|import { Badge } from 'antd'|import { Badge } from '@/components/ui/badge'|g
s|import { Tooltip } from 'antd'|import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from '@/components/ui/tooltip'|g
s|import { Tabs } from 'antd'|import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'|g
"

# 替换 @ant-design/icons 引用
find /home/shangmeilin/cube-castle/nextjs-app/src -name "*.tsx" -o -name "*.ts" | xargs sed -i "
s|import { ReloadOutlined } from '@ant-design/icons'|import { RefreshCw } from 'lucide-react'|g
s|import { ExclamationCircleOutlined } from '@ant-design/icons'|import { AlertTriangle } from 'lucide-react'|g
s|import { CheckCircleOutlined } from '@ant-design/icons'|import { CheckCircle } from 'lucide-react'|g
s|import { CloseCircleOutlined } from '@ant-design/icons'|import { XCircle } from 'lucide-react'|g
s|import { LoadingOutlined } from '@ant-design/icons'|import { Loader2 } from 'lucide-react'|g
s|import { SyncOutlined } from '@ant-design/icons'|import { RefreshCw } from 'lucide-react'|g
s|import { BugOutlined } from '@ant-design/icons'|import { Bug } from 'lucide-react'|g
s|import { ThunderboltOutlined } from '@ant-design/icons'|import { Zap } from 'lucide-react'|g
s|import { HomeOutlined } from '@ant-design/icons'|import { Home } from 'lucide-react'|g
s|import { PlusOutlined } from '@ant-design/icons'|import { Plus } from 'lucide-react'|g
s|import { EditOutlined } from '@ant-design/icons'|import { Edit } from 'lucide-react'|g
s|import { HistoryOutlined } from '@ant-design/icons'|import { History } from 'lucide-react'|g
s|import { UserOutlined } from '@ant-design/icons'|import { User } from 'lucide-react'|g
s|import { ArrowLeftOutlined } from '@ant-design/icons'|import { ArrowLeft } from 'lucide-react'|g
s|import { DatabaseOutlined } from '@ant-design/icons'|import { Database } from 'lucide-react'|g
s|import { WarningOutlined } from '@ant-design/icons'|import { AlertTriangle } from 'lucide-react'|g
"

# 替换组件使用
find /home/shangmeilin/cube-castle/nextjs-app/src -name "*.tsx" -o -name "*.ts" | xargs sed -i "
s|<ReloadOutlined />|<RefreshCw className=\"h-4 w-4\" />|g
s|<ExclamationCircleOutlined />|<AlertTriangle className=\"h-4 w-4\" />|g
s|<CheckCircleOutlined />|<CheckCircle className=\"h-4 w-4\" />|g
s|<CloseCircleOutlined />|<XCircle className=\"h-4 w-4\" />|g
s|<LoadingOutlined />|<Loader2 className=\"h-4 w-4 animate-spin\" />|g
s|<SyncOutlined />|<RefreshCw className=\"h-4 w-4\" />|g
s|<BugOutlined />|<Bug className=\"h-4 w-4\" />|g
s|<ThunderboltOutlined />|<Zap className=\"h-4 w-4\" />|g
s|<HomeOutlined />|<Home className=\"h-4 w-4\" />|g
s|<PlusOutlined />|<Plus className=\"h-4 w-4\" />|g
s|<EditOutlined />|<Edit className=\"h-4 w-4\" />|g
s|<HistoryOutlined />|<History className=\"h-4 w-4\" />|g
s|<UserOutlined />|<User className=\"h-4 w-4\" />|g
s|<ArrowLeftOutlined />|<ArrowLeft className=\"h-4 w-4\" />|g
s|<DatabaseOutlined />|<Database className=\"h-4 w-4\" />|g
s|<WarningOutlined />|<AlertTriangle className=\"h-4 w-4\" />|g
"

# 替换 notification 使用
find /home/shangmeilin/cube-castle/nextjs-app/src -name "*.tsx" -o -name "*.ts" | xargs sed -i "
s|notification\.success|toast.success|g
s|notification\.error|toast.error|g
s|notification\.warning|toast.error|g
s|notification\.info|toast|g
"

echo "批量替换完成！"