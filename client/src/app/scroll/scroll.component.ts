import { AfterViewInit, Component, ElementRef, OnDestroy, ViewChildren } from '@angular/core';
import { Subscription } from 'rxjs/Subscription';
import { Point } from './point';

// https://github.com/angular/angular/issues/8663
@Component({
  templateUrl: './scroll.component.html',
  styleUrls: ['./scroll.component.css'],
})
export class ScrollComponent implements AfterViewInit, OnDestroy {
  @ViewChildren('pointListEntry') pointListEntry: any;
  pointListEntryChangeSubscription: Subscription;
  pointList: Array<Point> = [];
  activePoint = 0;
  activePosition = 0.0;

  constructor(public elementRef: ElementRef) {
  }

  calculateActivePosition() {
    const scroller: HTMLElement = this.elementRef.nativeElement.querySelector('.scroller');
    const active = scroller.querySelector('.active-point');
    if (!active) {
      this.activePosition = 0;
    } else {
      this.activePosition = active.getBoundingClientRect().top - scroller.getBoundingClientRect().top + scroller.scrollTop;
    }
  }

  ngAfterViewInit() {
    // For the first time
    this.onPointListEntryChanged();

    // For next times
    this.pointListEntryChangeSubscription = this.pointListEntry.changes.subscribe(() => {
      this.onPointListEntryChanged();
    });
  }

  ngOnDestroy() {
    if (this.pointListEntryChangeSubscription) {
      this.pointListEntryChangeSubscription.unsubscribe();
      this.pointListEntryChangeSubscription = null;
    }
  }

  onPointListEntryChanged() {
    // 現在アクティブなセルにスクロールする
    // Angular 自体にはスクロールを制御する機能がないため、
    // DOM に直接アクセスして実現する。
    const scroller: HTMLElement = this.elementRef.nativeElement.querySelector('.scroller');
    const active = scroller.querySelector('.active-point');
    if (!active) {
      scroller.scrollTop = 0;
      return;
    }

    const activeRect = active.getBoundingClientRect();
    const scrollerRect = scroller.getBoundingClientRect();
    // scroller 内でのアクティブセルの相対位置
    // アクティブセルのページ内位置 - scroller のページ内位置 + 現在のスクロール位置
    const activeTop = activeRect.top - scrollerRect.top + scroller.scrollTop;

    // activeTop にスクロールするとアクティブセルが一番上に来る状態にスクロールする。
    // 真ん中に表示したいので、以下の補正を行う
    // + scroller の高さ / 2
    // - セルの高さ / 2
    const scrollTarget = activeTop - (scroller.clientHeight / 2) + ((activeRect.bottom - activeRect.top) / 2);

    scroller.scrollTop = scrollTarget;
  }

  generatePoints() {
    const newPointList: Array<Point> = [];
    const size = 30;
    const active = Math.floor(Math.random() * size) + 1;
    const now = new Date();
    for (let i = 0; i < size; ++i) {
      const point = new Point();
      point.point = i + 1;
      if (i <= active) {
        point.createdAt = new Date(now.getTime() + (i * 24 * 60 * 60));
      }
      newPointList.push(point);
    }
    this.pointList = newPointList;
    this.activePoint = active;
  }
}
