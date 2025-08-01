"use client"

import * as React from "react"
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table"
import { ChevronDown, ArrowUpDown, MoreHorizontal } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
// 暂时移除Radix UI的DropdownMenu，使用原生实现避免循环依赖
// import {
//   DropdownMenu,
//   DropdownMenuCheckboxItem,
//   DropdownMenuContent,
//   DropdownMenuItem,
//   DropdownMenuLabel,
//   DropdownMenuSeparator,
//   DropdownMenuTrigger,
// } from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  searchKey?: string
  searchPlaceholder?: string
}

// 原生列设置下拉菜单组件，避免Radix UI循环依赖
const NativeColumnDropdown: React.FC<{
  table: any
}> = ({ table }) => {
  const [isOpen, setIsOpen] = React.useState(false);
  const menuRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
    
    return undefined;
  }, [isOpen]);

  return (
    <div className="relative" ref={menuRef}>
      <Button 
        variant="outline" 
        className="ml-auto"
        onClick={() => setIsOpen(!isOpen)}
      >
        列设置 <ChevronDown className="ml-2 h-4 w-4" />
      </Button>
      {isOpen && (
        <div className="absolute right-0 top-full mt-1 w-48 rounded-md border bg-popover p-1 text-popover-foreground shadow-md z-50">
          {table
            .getAllColumns()
            .filter((column: any) => column.getCanHide())
            .map((column: any) => (
              <label
                key={column.id}
                className="flex items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={column.getIsVisible()}
                  onChange={(e) => column.toggleVisibility(e.target.checked)}
                  className="rounded border border-input"
                />
                <span className="capitalize">{column.id}</span>
              </label>
            ))}
        </div>
      )}
    </div>
  );
};

NativeColumnDropdown.displayName = 'NativeColumnDropdown';

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  searchPlaceholder = "搜索...",
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = React.useState({})

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  })

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        {searchKey && (
          <Input
            placeholder={searchPlaceholder}
            value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
            onChange={(event) =>
              table.getColumn(searchKey)?.setFilterValue(event.target.value)
            }
            className="max-w-sm"
          />
        )}
        <NativeColumnDropdown table={table} />
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          已选择 {table.getFilteredSelectedRowModel().rows.length} 项，共{" "}
          {table.getFilteredRowModel().rows.length} 项
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            上一页
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            下一页
          </Button>
        </div>
      </div>
    </div>
  )
}

// 通用列定义辅助函数
export function createSortableColumn<T>(
  accessorKey: keyof T,
  header: string
): ColumnDef<T> {
  return {
    accessorKey: accessorKey as string,
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          {header}
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
  }
}

export function createSelectColumn<T>(): ColumnDef<T> {
  return {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={table.getIsAllPageRowsSelected()}
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="选择全部"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="选择行"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  }
}

export function createActionsColumn<T>(
  actions: React.ComponentType<{ row: T }>
): ColumnDef<T> {
  const ActionsComponent = actions
  return {
    id: "actions",
    header: "操作",
    enableHiding: false,
    cell: ({ row }) => <ActionsComponent row={row.original} />,
  }
}

// 原生快速操作组件，避免Radix UI循环依赖
export function QuickActions<T>({
  row,
  actions,
}: {
  row: T
  actions: Array<{
    label: string
    onClick: (data: T) => void
    variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  }>
}) {
  const [isOpen, setIsOpen] = React.useState(false);
  const menuRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
    
    return undefined;
  }, [isOpen]);

  return (
    <div className="relative" ref={menuRef}>
      <Button 
        variant="ghost" 
        className="h-8 w-8 p-0"
        onClick={() => setIsOpen(!isOpen)}
      >
        <span className="sr-only">打开菜单</span>
        <MoreHorizontal className="h-4 w-4" />
      </Button>
      {isOpen && (
        <div className="absolute right-0 top-full mt-1 w-48 rounded-md border bg-popover p-1 text-popover-foreground shadow-md z-50">
          <div className="px-2 py-1.5 text-sm font-semibold text-foreground">操作</div>
          <div className="h-px bg-border my-1" />
          {actions.map((action, index) => (
            <button
              key={index}
              onClick={() => {
                action.onClick(row);
                setIsOpen(false);
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
            >
              {action.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}