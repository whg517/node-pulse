import { Fragment } from 'react'
import { Link } from 'react-router-dom'
import { useBreadcrumb } from './useBreadcrumb'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { House } from '@phosphor-icons/react'

export function AppBreadcrumb() {
  const { items } = useBreadcrumb()
  const lastIndex = items.length - 1

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {items.map((item, index) => {
          const isLast = index === lastIndex
          const isHome = index === 0

          return (
            <Fragment key={`breadcrumb-${index}`}>
              {index > 0 && <BreadcrumbSeparator />}
              <BreadcrumbItem>
                {isLast ? (
                  <BreadcrumbPage>
                    {isHome ? (
                      <House className="size-3.5" />
                    ) : (
                      item.label
                    )}
                  </BreadcrumbPage>
                ) : (
                  <BreadcrumbLink asChild>
                    <Link to={item.path}>
                      {isHome ? <House className="size-3.5" /> : item.label}
                    </Link>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            </Fragment>
          )
        })}
      </BreadcrumbList>
    </Breadcrumb>
  )
}

// Keep legacy export name for compatibility
export { AppBreadcrumb as Breadcrumb }
export default AppBreadcrumb
