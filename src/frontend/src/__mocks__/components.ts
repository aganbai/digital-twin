/**
 * Taro 组件 Mock
 * 用于单元测试中模拟 @tarojs/components
 */

import React, { ReactNode, HTMLAttributes, ButtonHTMLAttributes, ImgHTMLAttributes, InputHTMLAttributes, TextareaHTMLAttributes, FormHTMLAttributes, LabelHTMLAttributes } from 'react'

// Basic Taro components as div wrappers
export const View: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('div', { className, ...props }, children)

export const Text: React.FC<HTMLAttributes<HTMLSpanElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('span', { className, ...props }, children)

export const Button: React.FC<ButtonHTMLAttributes<HTMLButtonElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('button', { className, ...props }, children)

export const Image: React.FC<ImgHTMLAttributes<HTMLImageElement> & { className?: string; src?: string }> = ({
  className,
  src,
  ...props
}) => React.createElement('img', { className, src, alt: '', ...props })

export const Input: React.FC<InputHTMLAttributes<HTMLInputElement> & { className?: string }> = ({
  className,
  ...props
}) => React.createElement('input', { className, ...props })

export const Textarea: React.FC<TextareaHTMLAttributes<HTMLTextAreaElement> & { className?: string }> = ({
  className,
  ...props
}) => React.createElement('textarea', { className, ...props })

export const ScrollView: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; scrollY?: boolean; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('div', { className, ...props }, children)

export const Swiper: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('div', { className, ...props }, children)

export const SwiperItem: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('div', { className, ...props }, children)

export const Navigator: React.FC<HTMLAttributes<HTMLAnchorElement> & { className?: string; url?: string; children?: ReactNode }> = ({
  children,
  className,
  url,
  ...props
}) =>
  React.createElement('a', { className, href: url, ...props }, children)

export const Icon: React.FC<HTMLAttributes<HTMLSpanElement> & { className?: string; type?: string; size?: number | string }> = ({
  className,
  type,
  ...props
}) =>
  React.createElement('span', { className, 'data-type': type, ...props }, type)

export const Switch: React.FC<InputHTMLAttributes<HTMLInputElement> & { className?: string; checked?: boolean }> = ({
  className,
  checked,
  ...props
}) => React.createElement('input', { type: 'checkbox', className, checked, ...props })

export const Picker: React.FC<
  HTMLAttributes<HTMLDivElement> & {
    className?: string
    mode?: 'selector' | 'multiSelector' | 'time' | 'date' | 'region'
    value?: any
    range?: string[] | any[]
    onChange?: (e: { detail: { value: any } }) => void
    'data-testid'?: string
    children?: ReactNode
  }
> = ({ children, className, mode, range = [], onChange, ...props }) =>
  React.createElement(
    'div',
    { className, 'data-mode': mode, ...props },
    children,
    onChange && range.length > 0 &&
      React.createElement(
        'select',
        { 'data-testid': 'picker-select', onChange: (e) => onChange({ detail: { value: (e.target as HTMLSelectElement).value } }) },
        range.map((item, idx) =>
          React.createElement('option', { key: idx, value: idx }, typeof item === 'string' ? item : item.label || item)
        )
      )
  )

export const Form: React.FC<FormHTMLAttributes<HTMLFormElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('form', { className, ...props }, children)

export const Label: React.FC<LabelHTMLAttributes<HTMLLabelElement> & { className?: string; children?: ReactNode }> = ({
  children,
  className,
  ...props
}) =>
  React.createElement('label', { className, ...props }, children)

export const Checkbox: React.FC<InputHTMLAttributes<HTMLInputElement> & { className?: string; value?: string; checked?: boolean }> = ({
  className,
  value,
  checked,
  ...props
}) => React.createElement('input', { type: 'checkbox', className, value, checked, ...props })

export const Radio: React.FC<InputHTMLAttributes<HTMLInputElement> & { className?: string; value?: string; checked?: boolean }> = ({
  className,
  value,
  checked,
  ...props
}) => React.createElement('input', { type: 'radio', className, value, checked, ...props })

export const Progress: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; percent?: number }> = ({
  className,
  percent,
  ...props
}) =>
  React.createElement('div', { className, 'data-percent': percent, ...props }, React.createElement('progress', { value: percent, max: 100 }))

export const RichText: React.FC<HTMLAttributes<HTMLDivElement> & { className?: string; nodes?: string | any[] }> = ({
  className,
  nodes,
  ...props
}) =>
  React.createElement('div', { className, dangerouslySetInnerHTML: { __html: typeof nodes === 'string' ? nodes : '' }, ...props })
